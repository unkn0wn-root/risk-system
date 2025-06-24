// Package scontext provides a fluent builder for enriching Go contexts with structured data.
// It simplifies adding user information, request metadata, and custom fields to contexts.
package scontext

import (
	"context"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
	UserEmailKey contextKey = "user_mail"
	SessionIDKey contextKey = "session_id"
	UserRoleKey  contextKey = "user_role"
	UserRolesKey contextKey = "user_roles"
)

// Builder provides a fluent interface for building enriched contexts.
type Builder struct {
	ctx context.Context // The underlying context being enriched
}

// New creates a new context builder starting with the provided base context.
func New(ctx context.Context) *Builder {
	return &Builder{ctx: ctx}
}

// WithUserID adds a user ID to the context if the value is not empty.
func (b *Builder) WithUserID(userID string) *Builder {
	if userID != "" {
		b.ctx = context.WithValue(b.ctx, UserIDKey, userID)
	}
	return b
}

// WithUserEmail adds a user email to the context if the value is not empty.
func (b *Builder) WithUserEmail(email string) *Builder {
	if email != "" {
		b.ctx = context.WithValue(b.ctx, UserEmailKey, email)
	}
	return b
}

// WithUserRole adds a single user role to the context if the value is not empty.
func (b *Builder) WithUserRole(role string) *Builder {
	if role != "" {
		b.ctx = context.WithValue(b.ctx, UserRoleKey, role)
	}
	return b
}

// WithUserRoles adds multiple user roles to the context if the slice is not empty.
func (b *Builder) WithUserRoles(roles []string) *Builder {
	if len(roles) > 0 {
		b.ctx = context.WithValue(b.ctx, UserRolesKey, roles)
	}
	return b
}

// WithRequestID adds a request ID to the context for request tracing if the value is not empty.
func (b *Builder) WithRequestID(requestID string) *Builder {
	if requestID != "" {
		b.ctx = context.WithValue(b.ctx, RequestIDKey, requestID)
	}
	return b
}

// WithSessionID adds a session ID to the context for session tracking if the value is not empty.
func (b *Builder) WithSessionID(sessionID string) *Builder {
	if sessionID != "" {
		b.ctx = context.WithValue(b.ctx, SessionIDKey, sessionID)
	}
	return b
}

// WithCustomField adds a custom key-value pair to the context if both key and value are valid.
func (b *Builder) WithCustomField(key string, value any) *Builder {
	if key != "" && value != nil {
		b.ctx = context.WithValue(b.ctx, key, value)
	}
	return b
}

// Build returns the enriched context with all added fields.
func (b *Builder) Build() context.Context {
	return b.ctx
}

// WithUser is a convenience method that adds user ID, email, and role in one call.
func (b *Builder) WithUser(userID, email, role string) *Builder {
	return b.WithUserID(userID).WithUserEmail(email).WithUserRole(role)
}

// WithUserAndRoles is a convenience method that adds user ID, email, and multiple roles in one call.
func (b *Builder) WithUserAndRoles(userID, email string, roles []string) *Builder {
	return b.WithUserID(userID).WithUserEmail(email).WithUserRoles(roles)
}

// WithRequest is a convenience method that adds both request ID and session ID in one call.
func (b *Builder) WithRequest(requestID, sessionID string) *Builder {
	return b.WithRequestID(requestID).WithSessionID(sessionID)
}

// WithUserID creates a new builder from the context and adds a user ID.
func WithUserID(ctx context.Context, userID string) *Builder {
	return New(ctx).WithUserID(userID)
}

// WithUser creates a new builder from the context and adds user information.
func WithUser(ctx context.Context, userID, email, role string) *Builder {
	return New(ctx).WithUser(userID, email, role)
}

// WithRequest creates a new builder from the context and adds request tracking information.
func WithRequest(ctx context.Context, requestID, sessionID string) *Builder {
	return New(ctx).WithRequest(requestID, sessionID)
}
