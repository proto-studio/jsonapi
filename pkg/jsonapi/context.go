package jsonapi

import (
	"context"
)

type contextKey string

// WithMethod stores the HTTP method in the context for use by validators.
func WithMethod(ctx context.Context, method string) context.Context {
	return context.WithValue(ctx, contextKey("method"), method)
}

// MethodFromContext returns the HTTP method stored in the context, or empty string if unset.
func MethodFromContext(ctx context.Context) string {
	if s, ok := ctx.Value(contextKey("method")).(string); ok {
		return s
	}

	return ""
}

// WithId stores the resource ID in the context for use by validators.
func WithId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey("id"), id)
}

// IdFromContext returns the resource ID stored in the context, or empty string if unset.
func IdFromContext(ctx context.Context) string {
	if s, ok := ctx.Value(contextKey("id")).(string); ok {
		return s
	}

	return ""
}
