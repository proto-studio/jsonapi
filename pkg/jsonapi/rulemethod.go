package jsonapi

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for maximum
type httpMethodRule[T any, TK comparable] struct {
	methods []string
}

// Evaluate returns an error if the method in the context is not one of the specified methods.
// If no method is specified in the context, it will always return nil.
func (rule *httpMethodRule[T, TK]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
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
func (rule *httpMethodRule[T, TK]) Conflict(x rules.Rule[T]) bool {
	_, ok := x.(*httpMethodRule[T, TK])
	return ok
}

// String returns the string representation of the method rule.
func (rule *httpMethodRule[T, TK]) String() string {
	return fmt.Sprintf("HttpMethod(%v)", rule.methods)
}

// KeyRules returns an empty slice of rules.Rule[TK] since this rule is not key-specific.
// Implementing this method allows us to use this rule as a conditional rule in a ObjectRuleSet directly.
func (rule *httpMethodRule[T, TK]) KeyRules() []rules.Rule[TK] {
	return []rules.Rule[TK]{}
}

// HTTPMethodRule creates a new Rule that checks if the HTTP method is one of the specified methods.
func HTTPMethodRule[T any, TK comparable](methods ...string) *httpMethodRule[T, TK] {
	return &httpMethodRule[T, TK]{methods: methods}
}

// IndexRule creates a new Rule that checks if the request is an index request.
func IndexRule[T any, TK comparable]() rules.Rule[T] {
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
