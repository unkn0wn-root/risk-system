package scontext

import (
	"context"
)

type Builder struct {
	ctx context.Context
}

func New(ctx context.Context) *Builder {
	return &Builder{ctx: ctx}
}

func (b *Builder) WithUserID(userID string) *Builder {
	if userID != "" {
		b.ctx = context.WithValue(b.ctx, "user_id", userID)
	}
	return b
}

func (b *Builder) WithUserEmail(email string) *Builder {
	if email != "" {
		b.ctx = context.WithValue(b.ctx, "user_email", email)
	}
	return b
}

func (b *Builder) WithUserRole(role string) *Builder {
	if role != "" {
		b.ctx = context.WithValue(b.ctx, "user_role", role)
	}
	return b
}

func (b *Builder) WithUserRoles(roles []string) *Builder {
	if len(roles) > 0 {
		b.ctx = context.WithValue(b.ctx, "user_roles", roles)
	}
	return b
}

func (b *Builder) WithRequestID(requestID string) *Builder {
	if requestID != "" {
		b.ctx = context.WithValue(b.ctx, "request_id", requestID)
	}
	return b
}

func (b *Builder) WithSessionID(sessionID string) *Builder {
	if sessionID != "" {
		b.ctx = context.WithValue(b.ctx, "session_id", sessionID)
	}
	return b
}

func (b *Builder) WithCustomField(key string, value any) *Builder {
	if key != "" && value != nil {
		b.ctx = context.WithValue(b.ctx, key, value)
	}
	return b
}

func (b *Builder) Build() context.Context {
	return b.ctx
}

func (b *Builder) WithUser(userID, email, role string) *Builder {
	return b.WithUserID(userID).WithUserEmail(email).WithUserRole(role)
}

func (b *Builder) WithUserAndRoles(userID, email string, roles []string) *Builder {
	return b.WithUserID(userID).WithUserEmail(email).WithUserRoles(roles)
}

func (b *Builder) WithRequest(requestID, sessionID string) *Builder {
	return b.WithRequestID(requestID).WithSessionID(sessionID)
}

func WithUserID(ctx context.Context, userID string) *Builder {
	return New(ctx).WithUserID(userID)
}

func WithUser(ctx context.Context, userID, email, role string) *Builder {
	return New(ctx).WithUser(userID, email, role)
}

func WithRequest(ctx context.Context, requestID, sessionID string) *Builder {
	return New(ctx).WithRequest(requestID, sessionID)
}
