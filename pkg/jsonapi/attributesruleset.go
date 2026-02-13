package jsonapi

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// AttributesRuleSet wraps rules.ObjectRuleSet (StringMap[any]) for resource object attributes
// and adds JSON:API-safe key registration. WithKey panics if the key is not a valid
// member name per JSON:API (MemberNameRule); WithKeyUnsafe skips that check.
// All other ObjectRuleSet methods are delegated to the inner rule set.
type AttributesRuleSet struct {
	inner *rules.ObjectRuleSet[map[string]any, string, any]
}

// Attributes returns a new attributes rule set backed by rules.StringMap[any]().
func Attributes() *AttributesRuleSet {
	return &AttributesRuleSet{inner: rules.StringMap[any]()}
}

func (a *AttributesRuleSet) mustValidMemberName(name string) {
	rule := MemberNameRule{}
	if errs := rule.Evaluate(context.Background(), name); errs != nil {
		msg := "jsonapi: attribute name %q is not a valid JSON:API member name"
		if unwrapped := errors.Unwrap(errs); len(unwrapped) > 0 {
			msg = fmt.Sprintf("%s: %s", msg, unwrapped[0].Error())
		}
		panic(msg)
	}
}

// WithKey registers an attribute key and its rule set; panics if key is not a valid
// JSON:API member name (empty or contains reserved characters). Use MemberNameRule.Evaluate
// or WithKeyUnsafe to avoid panic when the key may be invalid.
func (a *AttributesRuleSet) WithKey(name string, ruleSet rules.RuleSet[any]) *AttributesRuleSet {
	a.mustValidMemberName(name)
	return &AttributesRuleSet{inner: a.inner.WithKey(name, ruleSet)}
}

// WithKeyUnsafe registers an attribute key without validating the key name.
func (a *AttributesRuleSet) WithKeyUnsafe(name string, ruleSet rules.RuleSet[any]) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithKey(name, ruleSet)}
}

// WithConditionalKey registers a conditional attribute key; panics if key is not a valid JSON:API member name.
func (a *AttributesRuleSet) WithConditionalKey(key string, condition rules.Conditional[map[string]any, string], ruleSet rules.RuleSet[any]) *AttributesRuleSet {
	a.mustValidMemberName(key)
	return &AttributesRuleSet{inner: a.inner.WithConditionalKey(key, condition, ruleSet)}
}

// WithConditionalKeyUnsafe registers a conditional attribute key without validating the key name.
func (a *AttributesRuleSet) WithConditionalKeyUnsafe(key string, condition rules.Conditional[map[string]any, string], ruleSet rules.RuleSet[any]) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithConditionalKey(key, condition, ruleSet)}
}

// WithDynamicKey adds a validation rule for any key that matches the key rule.
func (a *AttributesRuleSet) WithDynamicKey(keyRule rules.Rule[string], ruleSet rules.RuleSet[any]) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithDynamicKey(keyRule, ruleSet)}
}

// WithDynamicBucket puts matching keys into the named bucket (map key).
func (a *AttributesRuleSet) WithDynamicBucket(keyRule rules.Rule[string], bucket string) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithDynamicBucket(keyRule, bucket)}
}

// WithConditionalDynamicBucket puts matching keys into the bucket when the condition is met.
func (a *AttributesRuleSet) WithConditionalDynamicBucket(keyRule rules.Rule[string], condition rules.Conditional[map[string]any, string], bucket string) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithConditionalDynamicBucket(keyRule, condition, bucket)}
}

// KeyRules returns the key rules that have rule sets associated with them.
func (a *AttributesRuleSet) KeyRules() []rules.Rule[string] {
	return a.inner.KeyRules()
}

// WithUnknown allows any attribute key (dynamic attributes).
func (a *AttributesRuleSet) WithUnknown() *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithUnknown()}
}

// WithRequired returns a new rule set that requires the value to be present when nested.
func (a *AttributesRuleSet) WithRequired() *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithRequired()}
}

// WithJson allows the input to be a JSON-encoded string.
func (a *AttributesRuleSet) WithJson() *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithJson()}
}

// WithRule adds a custom validation rule over the entire attributes object.
func (a *AttributesRuleSet) WithRule(rule rules.Rule[map[string]any]) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithRule(rule)}
}

// WithRuleFunc adds a custom validation function over the entire attributes object.
func (a *AttributesRuleSet) WithRuleFunc(rule rules.RuleFunc[map[string]any]) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithRuleFunc(rule)}
}

// WithErrorMessage sets custom short and long error messages.
func (a *AttributesRuleSet) WithErrorMessage(short, long string) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithErrorMessage(short, long)}
}

// WithDocsURI sets a documentation URI on validation errors.
func (a *AttributesRuleSet) WithDocsURI(uri string) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithDocsURI(uri)}
}

// WithTraceURI sets a trace/debug URI on validation errors.
func (a *AttributesRuleSet) WithTraceURI(uri string) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithTraceURI(uri)}
}

// WithErrorCode overrides the error code for validation errors.
func (a *AttributesRuleSet) WithErrorCode(code errors.ErrorCode) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithErrorCode(code)}
}

// WithErrorMeta adds metadata to validation errors.
func (a *AttributesRuleSet) WithErrorMeta(key string, value any) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithErrorMeta(key, value)}
}

// WithErrorCallback sets a callback for custom error processing.
func (a *AttributesRuleSet) WithErrorCallback(fn errors.ErrorCallback) *AttributesRuleSet {
	return &AttributesRuleSet{inner: a.inner.WithErrorCallback(fn)}
}

// Apply implements rules.RuleSet[map[string]any].
func (a *AttributesRuleSet) Apply(ctx context.Context, input any) (map[string]any, errors.ValidationError) {
	return a.inner.Apply(ctx, input)
}

// Evaluate implements rules.RuleSet[map[string]any].
func (a *AttributesRuleSet) Evaluate(ctx context.Context, value map[string]any) errors.ValidationError {
	return a.inner.Evaluate(ctx, value)
}

// Required implements rules.RuleSet[map[string]any].
func (a *AttributesRuleSet) Required() bool {
	return a.inner.Required()
}

// String implements rules.RuleSet[map[string]any].
func (a *AttributesRuleSet) String() string {
	return a.inner.String()
}

// Replaces implements rules.RuleSet[map[string]any].
func (a *AttributesRuleSet) Replaces(r rules.Rule[map[string]any]) bool {
	return a.inner.Replaces(r)
}

// Any implements rules.RuleSet[map[string]any].
func (a *AttributesRuleSet) Any() rules.RuleSet[any] {
	return a.inner.Any()
}
