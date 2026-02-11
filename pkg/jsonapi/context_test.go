package jsonapi_test

import (
	"context"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
)

func TestWithMethod(t *testing.T) {
	ctx := context.Background()
	method := "GET"

	ctxWithMethod := jsonapi.WithMethod(ctx, method)
	retrievedMethod := jsonapi.MethodFromContext(ctxWithMethod)

	if retrievedMethod != method {
		t.Errorf("Expected method to be %q, got %q", method, retrievedMethod)
	}
}

func TestMethodFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	method := jsonapi.MethodFromContext(ctx)

	if method != "" {
		t.Errorf("Expected method to be empty string, got %q", method)
	}
}

func TestWithId(t *testing.T) {
	ctx := context.Background()
	id := "123"

	ctxWithId := jsonapi.WithId(ctx, id)
	retrievedId := jsonapi.IdFromContext(ctxWithId)

	if retrievedId != id {
		t.Errorf("Expected id to be %q, got %q", id, retrievedId)
	}
}

func TestIdFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	id := jsonapi.IdFromContext(ctx)

	if id != "" {
		t.Errorf("Expected id to be empty string, got %q", id)
	}
}

func TestContext_MethodAndId(t *testing.T) {
	ctx := context.Background()
	method := "POST"
	id := "456"

	ctx = jsonapi.WithMethod(ctx, method)
	ctx = jsonapi.WithId(ctx, id)

	retrievedMethod := jsonapi.MethodFromContext(ctx)
	retrievedId := jsonapi.IdFromContext(ctx)

	if retrievedMethod != method {
		t.Errorf("Expected method to be %q, got %q", method, retrievedMethod)
	}
	if retrievedId != id {
		t.Errorf("Expected id to be %q, got %q", id, retrievedId)
	}
}
