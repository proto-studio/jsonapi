package jsonapi

import (
	"context"
)

type contextKey string

func WithMethod(ctx context.Context, method string) context.Context {
	return context.WithValue(ctx, contextKey("method"), method)
}

func MethodFromContext(ctx context.Context) string {
	if s, ok := ctx.Value(contextKey("method")).(string); ok {
		return s
	}

	return ""
}

func WithId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey("id"), id)
}

func IdFromContext(ctx context.Context) string {
	if s, ok := ctx.Value(contextKey("id")).(string); ok {
		return s
	}

	return ""
}
