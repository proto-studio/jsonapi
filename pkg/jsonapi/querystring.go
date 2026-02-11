package jsonapi

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
	rulesnet "proto.zip/studio/validate/pkg/rules/net"
)

type SortParam struct {
	Field      string
	Descending bool
}

type ValueList interface {
	Values() []string
	Contains(field string) bool
	doNotExtend()
}

type fieldListMap map[string]bool

// Values returns the list of field names in the field list.
func (fl fieldListMap) Values() []string {
	keys := make([]string, 0, len(fl))
	for key := range fl {
		keys = append(keys, key)
	}
	return keys
}

// Contains reports whether the field list includes the given field name.
func (fl fieldListMap) Contains(field string) bool {
	v, ok := fl[field]
	return ok && v
}

// doNotExtend prevents external types from satisfying ValueList without the intended methods.
func (fieldListMap) doNotExtend() {}

// NewFieldList returns a ValueList containing the given field names.
func NewFieldList(fields ...string) ValueList {
	out := make(fieldListMap, len(fields))

	for _, field := range fields {
		out[field] = true
	}

	return out
}

type Include struct {
	Leaf string
	Path []string
}

// Standard JSON:API query parameter names that are all-lowercase (reserved by spec).
// Implementation-specific params must contain at least one non-lowercase character.
var jsonapiStandardLowercaseParams = map[string]bool{
	"sort": true, "include": true,
}

// QueryParamNameRule validates that a string is a valid JSON:API query parameter name per the spec.
// Implementation-specific names must contain at least one character outside a-z (e.g. uppercase, digit, bracket).
// Use with rules.String().WithRule(QueryParamNameRule) or Evaluate to check before calling WithParam.
type QueryParamNameRule struct{}

// Evaluate implements rules.Rule[string].
func (QueryParamNameRule) Evaluate(ctx context.Context, value string) errors.ValidationError {
	if isLegalQueryParamKey(value) {
		return nil
	}
	return errors.Errorf(errors.CodeUnexpected, ctx, "reserved query parameter", "query parameter name %q is reserved (all lowercase) for future JSON:API use", value)
}

// Replaces implements rules.Rule[string].
func (QueryParamNameRule) Replaces(r rules.Rule[string]) bool { return false }

// String implements rules.Rule[string].
func (QueryParamNameRule) String() string { return "QueryParamNameRule" }

var _ rules.Rule[string] = QueryParamNameRule{}

// isLegalQueryParamKey reports whether key is legal per JSON:API (implementation-specific
// params must not be all lowercase). WithParam panics if key is illegal.
func isLegalQueryParamKey(key string) bool {
	for _, r := range key {
		if r < 'a' || r > 'z' {
			return true // contains non-a-z → legal (implementation-specific)
		}
	}
	return jsonapiStandardLowercaseParams[key]
}

// QueryRuleSet wraps rules/net.QueryRuleSet and adds JSON:API-safe param registration.
// WithParam panics if the key is illegal per JSON:API (all-lowercase names are reserved).
type QueryRuleSet struct {
	inner *rulesnet.QueryRuleSet
}

// Query returns a new JSON:API query rule set backed by rules/net.Query().
func Query() *QueryRuleSet {
	return &QueryRuleSet{inner: rulesnet.Query()}
}

// WithParam registers a query parameter; panics if key is all-lowercase and not a
// standard JSON:API param (reserved for future spec use).
func (q *QueryRuleSet) WithParam(name string, ruleSet rules.RuleSet[any]) *QueryRuleSet {
	if !isLegalQueryParamKey(name) {
		panic("jsonapi: query parameter name \"" + name + "\" is illegal per JSON:API spec (all-lowercase names are reserved)")
	}
	return &QueryRuleSet{inner: q.inner.WithParam(name, ruleSet)}
}

// WithParamUnsafe registers a query parameter without checking key legality.
func (q *QueryRuleSet) WithParamUnsafe(name string, ruleSet rules.RuleSet[any]) *QueryRuleSet {
	return &QueryRuleSet{inner: q.inner.WithParam(name, ruleSet)}
}

// WithRule adds a validation rule over the entire query (url.Values).
func (q *QueryRuleSet) WithRule(rule rules.Rule[url.Values]) *QueryRuleSet {
	return &QueryRuleSet{inner: q.inner.WithRule(rule)}
}

// Apply implements rules.RuleSet[url.Values].
func (q *QueryRuleSet) Apply(ctx context.Context, input, output any) errors.ValidationError {
	return ToJSONAPIErrors(q.inner.Apply(ctx, input, output), SourceParameter)
}

// Evaluate implements rules.RuleSet[url.Values].
func (q *QueryRuleSet) Evaluate(ctx context.Context, values url.Values) errors.ValidationError {
	return ToJSONAPIErrors(q.inner.Evaluate(ctx, values), SourceParameter)
}

// Required implements rules.RuleSet[url.Values].
func (q *QueryRuleSet) Required() bool {
	return q.inner.Required()
}

// String implements rules.RuleSet[url.Values].
func (q *QueryRuleSet) String() string {
	return q.inner.String()
}

// Replaces implements rules.RuleSet[url.Values] by delegating to the inner rule set.
func (q *QueryRuleSet) Replaces(r rules.Rule[url.Values]) bool {
	return q.inner.Replaces(r)
}

// Any implements rules.RuleSet[url.Values].
func (q *QueryRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[url.Values](q)
}

var stringQueryValueRuleSet = rules.Slice[string]().WithItemRuleSet(rules.String()).WithMaxLen(1)
var intQueryValueRuleSet = rules.Slice[int]().WithItemRuleSet(rules.Int()).WithMaxLen(1)

var fieldKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^fields\[[^\]]+\]$`), "")
var filterKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^filter\[[^\]]+\]$`), "")

// Filter is only allowed on index GET requests
var filterRuleSet = rules.Slice[string]().WithItemRuleSet(rules.String()).WithRule(HTTPMethodRule[[]string, string]("GET", "HEAD")).WithRule(IndexRule[[]string, string]())

var fieldsRuleSet = rules.Interface[ValueList]().WithCast(func(ctx context.Context, value any) (ValueList, errors.ValidationError) {
	// Fields is allowed on all methods except DELETE
	method := MethodFromContext(ctx)

	if method == "DELETE" {
		return nil, errors.Errorf(errors.CodeForbidden, ctx, "Fields forbidden on DELETE", "Fields are not allowed on DELETE requests")
	}

	var strs []string
	verrs := stringQueryValueRuleSet.Apply(ctx, value, &strs)

	if verrs != nil {
		return nil, verrs
	}

	splitStrs := strings.Split(strs[0], ",")

	return NewFieldList(splitStrs...), nil
})

var includeRuleSet = fieldsRuleSet

var sortRuleSet = rules.Interface[[]SortParam]().WithCast(func(ctx context.Context, value any) ([]SortParam, errors.ValidationError) {

	// Sort is only allowed on index GET requests
	method := MethodFromContext(ctx)
	id := IdFromContext(ctx)

	if method != "" && (id != "" || (method != "GET" && method != "HEAD")) {
		return nil, errors.Errorf(errors.CodeForbidden, ctx, "Sort forbidden", "Sort is only allowed on index GET requests")
	}

	var strs []string
	verrs := stringQueryValueRuleSet.Apply(ctx, value, &strs)

	if verrs != nil {
		return nil, verrs
	}

	itms := strings.Split(strs[0], ",")

	out := make([]SortParam, len(itms))

	for i, itm := range itms {
		if len(itm) < 1 {
			continue
		}

		if itm[0] == '-' {
			out[i] = SortParam{
				Field:      itm[1:],
				Descending: true,
			}
		} else {
			out[i] = SortParam{
				Field: itm,
			}
		}
	}

	return out, nil
})

var pageSizeRuleSet = intQueryValueRuleSet.WithRule(HTTPMethodRule[[]int, string]("GET", "HEAD")).WithRule(IndexRule[[]int, string]()).WithItemRuleSet(rules.Int().WithMin(1).WithMax(100)).Any()

var cursorRuleSet = rules.Slice[string]().WithItemRuleSet(rules.String().WithMinLen(1)).WithMaxLen(1).WithMinLen(1).WithRule(HTTPMethodRule[[]string, string]("GET", "HEAD")).WithRule(IndexRule[[]string, string]()).Any()

// jsonAPIQueryRule validates dynamic keys (fields[*], filter[*], ext) and rejects unknown all-lowercase params.
func jsonAPIQueryRule(ctx context.Context, values url.Values) errors.ValidationError {
	var allErrors []error
	for key, v := range values {
		paramCtx := rulecontext.WithPathString(ctx, "query["+key+"]")
		if fieldKeyRule.Evaluate(ctx, key) == nil {
			var fl ValueList
			if errs := fieldsRuleSet.Apply(paramCtx, v, &fl); errs != nil {
				allErrors = append(allErrors, errors.Unwrap(errs)...)
			}
			continue
		}
		if filterKeyRule.Evaluate(ctx, key) == nil {
			var sl []string
			if errs := filterRuleSet.Apply(paramCtx, v, &sl); errs != nil {
				allErrors = append(allErrors, errors.Unwrap(errs)...)
			}
			continue
		}
		if extKeyRule.Evaluate(ctx, key) == nil {
			continue
		}
		if !isLegalQueryParamKey(key) {
			allErrors = append(allErrors, errors.Errorf(errors.CodeUnexpected, paramCtx, "reserved query parameter", "query parameter %q is reserved (all lowercase) for future JSON:API use", key))
		}
	}
	return errors.Join(allErrors...)
}

// queryParamAdapter adapts rule sets that expect []string (from url.Values) to accept
// a single string (as passed by rules/net.QueryRuleSet per param).
type queryParamAdapter struct {
	inner rules.RuleSet[any]
}

// Apply converts a single string value to []string and delegates to the inner rule set.
func (a *queryParamAdapter) Apply(ctx context.Context, value, output any) errors.ValidationError {
	if s, ok := value.(string); ok {
		value = []string{s}
	}
	return a.inner.Apply(ctx, value, output)
}

// Evaluate converts a single string value to []string and delegates to the inner rule set.
func (a *queryParamAdapter) Evaluate(ctx context.Context, value any) errors.ValidationError {
	if s, ok := value.(string); ok {
		value = []string{s}
	}
	return a.inner.Evaluate(ctx, value)
}

// Required delegates to the inner rule set.
func (a *queryParamAdapter) Required() bool { return a.inner.Required() }

// String delegates to the inner rule set.
func (a *queryParamAdapter) String() string { return a.inner.String() }

// Replaces delegates to the inner rule set.
func (a *queryParamAdapter) Replaces(r rules.Rule[any]) bool { return a.inner.Replaces(r) }

// Any returns the adapter as rules.RuleSet[any] (the adapter itself).
func (a *queryParamAdapter) Any() rules.RuleSet[any] { return a }

// QueryStringBaseRuleSet is the default JSON:API query rule set. Use Apply to validate
// and coerce input (string or url.Values) to url.Values.
var QueryStringBaseRuleSet *QueryRuleSet = Query().
	WithParam("sort", &queryParamAdapter{inner: sortRuleSet.Any()}).
	WithParam("include", &queryParamAdapter{inner: includeRuleSet.Any()}).
	WithParamUnsafe("page[size]", &queryParamAdapter{inner: pageSizeRuleSet}).
	WithParamUnsafe("page[after]", &queryParamAdapter{inner: cursorRuleSet}).
	WithParamUnsafe("page[before]", &queryParamAdapter{inner: cursorRuleSet}).
	WithRule(rules.RuleFunc[url.Values](jsonAPIQueryRule))
