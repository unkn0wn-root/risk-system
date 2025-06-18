// Package scontext provides a fluent builder for enriching Go contexts with structured data.
// It simplifies adding user information, request metadata, and custom fields to contexts.
package scontext

import (
	"context"
)

// Builder provides a fluent interface for building enriched contexts.
// It allows chaining method calls to add various types of context data.
type Builder struct {
	ctx context.Context // The underlying context being enriched
}

// New creates a new context builder starting with the provided base context.
func New(ctx context.Context) *Builder {
	return &Builder{ctx: ctx}
}

// WithUserID adds a user ID to the context if the value is not empty.
// Returns the builder for method chaining.
func (b *Builder) WithUserID(userID string) *Builder {
	if userID != "" {
		b.ctx = context.WithValue(b.ctx, "user_id", userID)
	}
	return b
}

// WithUserEmail adds a user email to the context if the value is not empty.
// Returns the builder for method chaining.
func (b *Builder) WithUserEmail(email string) *Builder {
	if email != "" {
		b.ctx = context.WithValue(b.ctx, "user_email", email)
	}
	return b
}

// WithUserRole adds a single user role to the context if the value is not empty.
// Returns the builder for method chaining.
func (b *Builder) WithUserRole(role string) *Builder {
	if role != "" {
		b.ctx = context.WithValue(b.ctx, "user_role", role)
	}
	return b
}

// WithUserRoles adds multiple user roles to the context if the slice is not empty.
// Returns the builder for method chaining.
func (b *Builder) WithUserRoles(roles []string) *Builder {
	if len(roles) > 0 {
		b.ctx = context.WithValue(b.ctx, "user_roles", roles)
	}
	return b
}

// WithRequestID adds a request ID to the context for request tracing if the value is not empty.
// Returns the builder for method chaining.
func (b *Builder) WithRequestID(requestID string) *Builder {
	if requestID != "" {
		b.ctx = context.WithValue(b.ctx, "request_id", requestID)
	}
	return b
}

// WithSessionID adds a session ID to the context for session tracking if the value is not empty.
// Returns the builder for method chaining.
func (b *Builder) WithSessionID(sessionID string) *Builder {
	if sessionID != "" {
		b.ctx = context.WithValue(b.ctx, "session_id", sessionID)
	}
	return b
}

// WithCustomField adds a custom key-value pair to the context if both key and value are valid.
// Returns the builder for method chaining.
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
// Returns the builder for method chaining.
func (b *Builder) WithUser(userID, email, role string) *Builder {
	return b.WithUserID(userID).WithUserEmail(email).WithUserRole(role)
}

// WithUserAndRoles is a convenience method that adds user ID, email, and multiple roles in one call.
// Returns the builder for method chaining.
func (b *Builder) WithUserAndRoles(userID, email string, roles []string) *Builder {
	return b.WithUserID(userID).WithUserEmail(email).WithUserRoles(roles)
}

// WithRequest is a convenience method that adds both request ID and session ID in one call.
// Returns the builder for method chaining.
func (b *Builder) WithRequest(requestID, sessionID string) *Builder {
	return b.WithRequestID(requestID).WithSessionID(sessionID)
}

// WithUserID creates a new builder from the context and adds a user ID.
// This is a convenience function for single-field additions.
func WithUserID(ctx context.Context, userID string) *Builder {
	return New(ctx).WithUserID(userID)
}

// WithUser creates a new builder from the context and adds user information.
// This is a convenience function for adding user ID, email, and role together.
func WithUser(ctx context.Context, userID, email, role string) *Builder {
	return New(ctx).WithUser(userID, email, role)
}

// WithRequest creates a new builder from the context and adds request tracking information.
// This is a convenience function for adding request ID and session ID together.
func WithRequest(ctx context.Context, requestID, sessionID string) *Builder {
	return New(ctx).WithRequest(requestID, sessionID)
}
