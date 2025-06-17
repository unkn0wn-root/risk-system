package handlers

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	user_models "user-risk-system/cmd/user/models"
	"user-risk-system/cmd/user/repository"
	"user-risk-system/pkg/auth"
	"user-risk-system/pkg/errors"
	"user-risk-system/pkg/logger"
	"user-risk-system/pkg/messaging"
	"user-risk-system/pkg/models"
	pb_notification "user-risk-system/pkg/proto/notification"
	pb_risk "user-risk-system/pkg/proto/risk"
	pb_user "user-risk-system/pkg/proto/user"
	"user-risk-system/pkg/scontext"
)

type UserHandler struct {
	pb_user.UnimplementedUserServiceServer
	userRepo           *repository.UserRepository
	riskClient         pb_risk.RiskServiceClient
	notificationClient pb_notification.NotificationServiceClient
	messageQueue       *messaging.RabbitMQ
	logger             *logger.Logger
}

func NewUserHandler(
	userRepo *repository.UserRepository,
	riskClient pb_risk.RiskServiceClient,
	notificationClient pb_notification.NotificationServiceClient,
	messageQueue *messaging.RabbitMQ,
	appLogger *logger.Logger,
) *UserHandler {
	return &UserHandler{
		userRepo:           userRepo,
		riskClient:         riskClient,
		notificationClient: notificationClient,
		messageQueue:       messageQueue,
		logger:             appLogger,
	}
}

func (h *UserHandler) Login(ctx context.Context, req *pb_user.LoginRequest) (*pb_user.LoginResponse, error) {
	ctx = scontext.New(ctx).WithUserEmail(req.Email).Build()
	h.logger.InfoCtx(ctx, "Login attempt for email")

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		h.logger.ErrorCtx(ctx, "User not found", nil)
		return nil, errors.ErrUserNotFound.GRPCStatus().Err()
	}

	if !user.IsActive {
		h.logger.ErrorCtx(ctx, "Inactive user login attempt", nil)
		inactiveErr := &errors.AppError{
			Code:    "USER_INACTIVE",
			Message: "Account is deactivated",
		}
		return nil, inactiveErr.GRPCStatus().Err()
	}

	if !user.CheckPassword(req.Password) {
		h.logger.ErrorCtx(ctx, "Invalid password for user", nil)
		return nil, errors.ErrInvalidPassword.GRPCStatus().Err()
	}

	go h.checkLoginRisk(user)

	now := time.Now()
	user.LastLoginAt = &now
	if err := h.userRepo.Update(user); err != nil {
		// Don't fail login for this, just log it
		h.logger.ErrorCtx(ctx, "Failed to update last login time", err)
	}

	h.logger.InfoCtx(ctx, "Successful login")

	pbUser := h.userToProto(user)
	return &pb_user.LoginResponse{
		User: pbUser,
	}, nil
}

func (h *UserHandler) Register(ctx context.Context, req *pb_user.RegisterRequest) (*pb_user.RegisterResponse, error) {
	ctx = scontext.New(ctx).WithUserEmail(req.Email).Build()
	h.logger.InfoCtx(ctx, "Registration attempt for email")

	existingUser, _ := h.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		h.logger.ErrorCtx(ctx, "User already exists", nil)
		return nil, errors.ErrEmailExists.GRPCStatus().Err()
	}

	user := &user_models.User{
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Phone:      req.Phone,
		Roles:      []string{string(auth.RoleUser)}, // Default role
		IsActive:   true,
		IsVerified: false,
		CreatedAt:  time.Now(),
	}

	if err := user.SetPassword(req.Password); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to hash password", err)
		passErr := errors.ErrPasswordHashFailed.WithDetails(err.Error())
		return nil, passErr.GRPCStatus().Err()
	}

	if err := h.userRepo.Create(user); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to create user", err)
		createErr := errors.ErrUserCreateFailed.WithDetails(err.Error())
		return nil, createErr.GRPCStatus().Err()
	}

	ctxWithUserId := scontext.WithUserID(ctx, user.ID).Build()
	h.logger.InfoCtx(ctxWithUserId, "User registered successfully", "user_id", user.ID)

	pbUser := h.userToProto(user)

	go h.handleUserCreatedAsync(user)
	go h.handleUserCreatedSync(user)

	return &pb_user.RegisterResponse{
		User: pbUser,
	}, nil
}

// CreateUser creates a new user (admin function)
func (h *UserHandler) CreateUser(ctx context.Context, req *pb_user.CreateUserRequest) (*pb_user.CreateUserResponse, error) {
	ctx = scontext.New(ctx).WithUserEmail(req.Email).Build()
	h.logger.InfoCtx(ctx, "Admin creating user")

	// Check if user already exists
	existingUser, _ := h.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		h.logger.ErrorCtx(ctx, "User already exists", nil)
		return nil, errors.ErrEmailExists.GRPCStatus().Err()
	}

	user := &user_models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Roles:     []string{string(auth.RoleUser)}, // Default role
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := h.userRepo.Create(user); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to create user", err)
		createErr := errors.ErrUserCreateFailed.WithDetails(err.Error())
		return nil, createErr.GRPCStatus().Err()
	}

	ctxWithUserId := scontext.WithUserID(ctx, user.ID).Build()
	h.logger.InfoCtx(ctxWithUserId, "User created successfully by admin")

	pbUser := h.userToProto(user)

	go h.handleUserCreatedAsync(user)
	go h.handleUserCreatedSync(user)

	return &pb_user.CreateUserResponse{
		User: pbUser,
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb_user.GetUserRequest) (*pb_user.GetUserResponse, error) {
	userID := ctx.Value("user_id").(string)
	userRoles := ctx.Value("user_roles").([]string)

	isAdmin := false
	for _, role := range userRoles {
		if role == string(auth.RoleAdmin) {
			isAdmin = true
			break
		}
	}

	// Users can only access their own data unless they're admin
	if req.Id != userID && !isAdmin {
		return nil, errors.ErrInsufficientRole.GRPCStatus().Err()
	}

	user, err := h.userRepo.GetByID(req.Id)
	if err != nil {
		return nil, errors.ErrUserNotFound.GRPCStatus().Err()
	}

	pbUser := h.userToProto(user)
	return &pb_user.GetUserResponse{
		User: pbUser,
	}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *pb_user.UpdateUserRequest) (*pb_user.UpdateUserResponse, error) {
	userID := ctx.Value("user_id").(string)
	userRoles := ctx.Value("user_roles").([]string)

	isAdmin := false
	for _, role := range userRoles {
		if role == string(auth.RoleAdmin) {
			isAdmin = true
			break
		}
	}

	if req.Id != userID && !isAdmin {
		return nil, errors.ErrInsufficientRole.GRPCStatus().Err()
	}

	user, err := h.userRepo.GetByID(req.Id)
	if err != nil {
		return nil, errors.ErrUserNotFound.GRPCStatus().Err()
	}

	// Update user fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := h.userRepo.Update(user); err != nil {
		updateErr := errors.ErrUserUpdateFailed.WithDetails(err.Error())
		return nil, updateErr.GRPCStatus().Err()
	}

	pbUser := h.userToProto(user)

	return &pb_user.UpdateUserResponse{
		User: pbUser,
	}, nil
}

// Helper function to convert user model to protobuf
func (h *UserHandler) userToProto(user *user_models.User) *pb_user.User {
	pbUser := &pb_user.User{
		Id:         user.ID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Phone:      user.Phone,
		Roles:      user.Roles,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		CreatedAt:  timestamppb.New(user.CreatedAt),
	}

	if user.LastLoginAt != nil {
		pbUser.LastLoginAt = timestamppb.New(*user.LastLoginAt)
	}

	return pbUser
}

// handleUserCreatedAsync async gRPC queue
func (h *UserHandler) handleUserCreatedAsync(user *user_models.User) {
	event := models.UserCreatedEvent{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
	}

	if err := h.messageQueue.Publish("user.created", event); err != nil {
		h.logger.Error("Failed to publish user created event", err)
	}
}

// handleUserCreatedSync - via grpc
func (h *UserHandler) handleUserCreatedSync(user *user_models.User) {
	riskReq := &pb_risk.RiskCheckRequest{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
	}

	ctx := context.Background()
	ctx = scontext.New(ctx).WithUserAndRoles(user.ID, user.Email, user.Roles).Build()

	riskResp, err := h.riskClient.CheckRisk(ctx, riskReq)
	if err != nil {
		h.logger.ErrorCtx(ctx, "Failed to check risk for user", err)
		return
	}

	notificationReq := &pb_notification.SendNotificationRequest{
		UserId:  user.ID,
		Type:    "USER_CREATED",
		Message: "Welcome! Your account has been created successfully.",
		Email:   user.Email,
	}

	_, err = h.notificationClient.SendNotification(ctx, notificationReq)
	if err != nil {
		h.logger.ErrorCtx(ctx, "Failed to send user created notification", err)
	}

	if riskResp.IsRisky {
		var action string
		switch riskResp.RiskLevel {
		case "CRITICAL":
			action = "Account flagged for immediate review"
			go h.handleCriticalRisk(user, riskResp)
		case "HIGH":
			action = "Account requires verification"
			go h.handleHighRisk(user, riskResp)
		case "MEDIUM":
			action = "Account flagged for monitoring"
		default:
			action = "Low risk detected"
		}

		riskNotificationReq := &pb_notification.SendNotificationRequest{
			UserId:  user.ID,
			Type:    "RISK_DETECTED",
			Message: fmt.Sprintf("Risk detected (%s): %s. Action: %s", riskResp.RiskLevel, riskResp.Reason, action),
			Email:   user.Email,
		}

		_, err = h.notificationClient.SendNotification(ctx, riskNotificationReq)
		if err != nil {
			h.logger.ErrorCtx(ctx, "Failed to send risk notification", err)
		}

		h.logger.InfoCtx(ctx, "Risk detected for user",
			"risk_level", riskResp.RiskLevel,
			"reason", riskResp.Reason,
			"flags", riskResp.Flags,
			"action", action,
		)
	}
}

func (h *UserHandler) handleCriticalRisk(user *user_models.User, riskResp *pb_risk.RiskCheckResponse) {
	ctx := context.Background()
	ctx = scontext.New(ctx).WithUserID(user.ID).WithUserEmail(user.Email).Build()

	// Could automatically deactivate account
	user.IsActive = false
	if err := h.userRepo.Update(user); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to deactivate high-risk user", err)
	}

	adminAlert := &pb_notification.SendNotificationRequest{
		UserId:  "admin",
		Type:    "CRITICAL_RISK_ALERT",
		Message: fmt.Sprintf("CRITICAL RISK USER: %s (%s) - %s", user.Email, user.ID, riskResp.Reason),
		Email:   "admin@fakeasfake.com",
	}

	h.notificationClient.SendNotification(ctx, adminAlert)
}

func (h *UserHandler) handleHighRisk(user *user_models.User, riskResp *pb_risk.RiskCheckResponse) {
	ctx := context.Background()
	ctx = scontext.New(ctx).WithUserID(user.ID).WithUserEmail(user.Email).Build()

	user.IsVerified = false
	if err := h.userRepo.Update(user); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to update high-risk user verification", err)
	}

	// Send verification email
	verificationReq := &pb_notification.SendNotificationRequest{
		UserId:  user.ID,
		Type:    "EMAIL_VERIFICATION_REQUIRED",
		Message: "Please verify your email address to complete your account setup.",
		Email:   user.Email,
	}

	h.notificationClient.SendNotification(ctx, verificationReq)
}

func (h *UserHandler) checkLoginRisk(user *user_models.User) {
	ctx := context.Background()
	ctx = scontext.New(ctx).WithUserID(user.ID).WithUserEmail(user.Email).Build()

	// Could check for suspicious login patterns
	riskReq := &pb_risk.RiskCheckRequest{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
	}

	riskResp, err := h.riskClient.CheckRisk(ctx, riskReq)
	if err != nil {
		h.logger.ErrorCtx(ctx, "Failed to check login risk", err)
		return
	}

	if riskResp.IsRisky && riskResp.RiskLevel == "CRITICAL" {
		loginAlert := &pb_notification.SendNotificationRequest{
			UserId:  user.ID,
			Type:    "SUSPICIOUS_LOGIN_ALERT",
			Message: "Suspicious login detected on your account.",
			Email:   user.Email,
		}

		h.notificationClient.SendNotification(ctx, loginAlert)
	}
}
