package jsonapi

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for maximum
type httpMethodRule[T any] struct {
	methods []string
}

// Evaluate returns an error if the method in the context is not one of the specified methods.
// If no method is specified in the context, it will always return nil.
func (rule *httpMethodRule[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	method := MethodFromContext(ctx)

	for _, allowedMethod := range rule.methods {
		if method == allowedMethod {
			return nil
		}
	}

	return errors.Collection(
		errors.Errorf(errors.CodePattern, ctx, "HTTP method must be one of %v", rule.methods),
	)
}

// Conflict returns true for any maximum rule.
func (rule *httpMethodRule[T]) Conflict(x rules.Rule[T]) bool {
	_, ok := x.(*httpMethodRule[T])
	return ok
}

// String returns the string representation of the method rule.
func (rule *httpMethodRule[T]) String() string {
	return fmt.Sprintf("HttpMethod(%v)", rule.methods)
}

func HTTPMethodRule[T any](methods ...string) rules.Rule[T] {
	return &httpMethodRule[T]{methods: methods}
}

func IndexRule[T any]() rules.Rule[T] {
	return rules.RuleFunc[T](func(ctx context.Context, value T) errors.ValidationErrorCollection {
		method := MethodFromContext(ctx)
		id := IdFromContext(ctx)

		if method != "" && id != "" {
			return errors.Collection(
				errors.Errorf(errors.CodeForbidden, ctx, "Value is only allowed on index requests"),
			)
		}

		return nil
	})
}
