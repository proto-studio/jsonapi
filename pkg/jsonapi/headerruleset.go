package jsonapi

import (
	"context"
	"mime"
	"net/http"
	"reflect"
	"strings"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// MediaTypeJSONAPI is the JSON:API media type for Content-Type (per https://jsonapi.org/format/).
const MediaTypeJSONAPI = "application/vnd.api+json"

// Only these Content-Type parameters are allowed by the JSON:API spec.
const (
	contentTypeParamExt     = "ext"
	contentTypeParamProfile = "profile"
)

// HeaderRuleSet validates HTTP headers per JSON:API (Content-Type and optional ext/profile, plus custom headers).
// At minimum it checks that Content-Type is application/vnd.api+json with no disallowed parameters.
// Use WithExt/WithProfile to validate the ext and profile media type parameters; use WithHeader to validate other headers.
type HeaderRuleSet struct {
	contentRequired bool
	extRuleSet      rules.RuleSet[any]
	profileRuleSet  rules.RuleSet[any]
	headerRules     map[string]rules.RuleSet[any]
}

// Headers returns a new HeaderRuleSet that validates Content-Type and optionally ext/profile and custom headers.
func Headers() *HeaderRuleSet {
	return &HeaderRuleSet{
		contentRequired: true,
		headerRules:     make(map[string]rules.RuleSet[any]),
	}
}

func (h *HeaderRuleSet) clone() *HeaderRuleSet {
	c := &HeaderRuleSet{
		contentRequired: h.contentRequired,
		extRuleSet:      h.extRuleSet,
		profileRuleSet:  h.profileRuleSet,
		headerRules:     make(map[string]rules.RuleSet[any]),
	}
	for k, v := range h.headerRules {
		c.headerRules[k] = v
	}
	return c
}

// WithContentRequired sets whether Content-Type is required (default true).
func (h *HeaderRuleSet) WithContentRequired(required bool) *HeaderRuleSet {
	c := h.clone()
	c.contentRequired = required
	return c
}

// WithExt validates the Content-Type ext parameter value with the given rule set.
// The value is the raw ext parameter (e.g. space-separated URIs per JSON:API).
func (h *HeaderRuleSet) WithExt(ruleSet rules.RuleSet[any]) *HeaderRuleSet {
	c := h.clone()
	c.extRuleSet = ruleSet
	return c
}

// WithProfile validates the Content-Type profile parameter value with the given rule set.
// The value is the raw profile parameter (e.g. space-separated URIs per JSON:API).
func (h *HeaderRuleSet) WithProfile(ruleSet rules.RuleSet[any]) *HeaderRuleSet {
	c := h.clone()
	c.profileRuleSet = ruleSet
	return c
}

// WithHeader registers validation for the given header name (e.g. "Content-Type", "Accept").
// The rule set receives the first value of the header. Like WithKey on an object rule set.
func (h *HeaderRuleSet) WithHeader(name string, ruleSet rules.RuleSet[any]) *HeaderRuleSet {
	c := h.clone()
	if c.headerRules == nil {
		c.headerRules = make(map[string]rules.RuleSet[any])
	}
	c.headerRules[name] = ruleSet
	return c
}

// getHeader returns the first value for name from headers. Name is case-insensitive under http.Header.
func getHeader(headers http.Header, name string) string {
	v := headers[name]
	if len(v) == 0 {
		return ""
	}
	return strings.TrimSpace(v[0])
}

// headerToHTTP converts the JSON:API Header struct into http.Header for validation.
// Builds Content-Type from Version, Ext, and Profile (ext/profile params are space-separated URIs per JSON:API).
func headerToHTTP(in *Header) http.Header {
	h := make(http.Header)
	ct := MediaTypeJSONAPI
	if in != nil {
		if len(in.Ext) > 0 {
			uris := make([]string, len(in.Ext))
			for i, e := range in.Ext {
				uris[i] = e.URI
			}
			ct += "; ext=\"" + strings.Join(uris, " ") + "\""
		}
		if len(in.Profile) > 0 {
			uris := make([]string, len(in.Profile))
			for i, p := range in.Profile {
				uris[i] = p.URI
			}
			ct += "; profile=\"" + strings.Join(uris, " ") + "\""
		}
	}
	h.Set("Content-Type", ct)
	return h
}

// httpHeaderToHeader parses Content-Type and populates the JSON:API Header struct (Version, Ext, Profile).
// Meta is not set from headers. Version defaults to Version_1_1 when Content-Type is present.
func httpHeaderToHeader(headers http.Header) *Header {
	out := &Header{Version: Version_1_1}
	raw := getHeader(headers, "Content-Type")
	if raw == "" {
		return out
	}
	_, params, err := mime.ParseMediaType(raw)
	if err != nil {
		return out
	}
	if v := params[contentTypeParamExt]; v != "" {
		for _, uri := range strings.Fields(v) {
			out.Ext = append(out.Ext, Extension{URI: uri})
		}
	}
	if v := params[contentTypeParamProfile]; v != "" {
		for _, uri := range strings.Fields(v) {
			out.Profile = append(out.Profile, Profile{URI: uri})
		}
	}
	return out
}

// validateContentType checks Content-Type is application/vnd.api+json and only ext/profile params.
func (h *HeaderRuleSet) validateContentType(ctx context.Context, headers http.Header) errors.ValidationError {
	headerCtx := rulecontext.WithPathString(ctx, "Content-Type")
	raw := getHeader(headers, "Content-Type")
	if raw == "" {
		if h.contentRequired {
			return errors.Errorf(errors.CodeRequired, headerCtx, "Content-Type required", "Content-Type header must be %s", MediaTypeJSONAPI)
		}
		return nil
	}
	mediaType, params, err := mime.ParseMediaType(raw)
	if err != nil {
		return errors.Errorf(errors.CodeEncoding, headerCtx, "invalid Content-Type", "Content-Type header is invalid: %v", err)
	}
	if mediaType != MediaTypeJSONAPI {
		return errors.Errorf(errors.CodePattern, headerCtx, "wrong media type", "Content-Type must be %s, got %q", MediaTypeJSONAPI, mediaType)
	}
	for name := range params {
		if name != contentTypeParamExt && name != contentTypeParamProfile {
			return errors.Errorf(errors.CodeUnexpected, headerCtx, "disallowed parameter", "Content-Type parameter %q is not allowed (only ext and profile)", name)
		}
	}

	// Validate ext parameter value if rule set configured
	if h.extRuleSet != nil {
		if extVal := params[contentTypeParamExt]; extVal != "" {
			extCtx := rulecontext.WithPathString(ctx, "Content-Type")
			if _, err := h.extRuleSet.Apply(extCtx, extVal); err != nil {
				return err
			}
		}
	}
	// Validate profile parameter value if rule set configured
	if h.profileRuleSet != nil {
		if profileVal := params[contentTypeParamProfile]; profileVal != "" {
			profileCtx := rulecontext.WithPathString(ctx, "Content-Type")
			if _, err := h.profileRuleSet.Apply(profileCtx, profileVal); err != nil {
				return err
			}
		}
	}
	return nil
}

// Evaluate validates headers. Per JSON:API, Content-Type must be application/vnd.api+json with only ext/profile params.
func (h *HeaderRuleSet) Evaluate(ctx context.Context, headers http.Header) errors.ValidationError {
	var errs []error
	contentErr := h.validateContentType(ctx, headers)
	if contentErr != nil {
		errs = append(errs, errors.Unwrap(contentErr)...)
	}
	for name, ruleSet := range h.headerRules {
		if ruleSet == nil {
			continue
		}
		val := getHeader(headers, name)
		headerCtx := rulecontext.WithPathString(ctx, name)
		if _, err := ruleSet.Apply(headerCtx, val); err != nil {
			errs = append(errs, errors.Unwrap(err)...)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return ToJSONAPIErrors(errors.Join(errs...), SourceHeader)
}

// Apply coerces input to http.Header (or from map[string][]string or jsonapi Header), validates, and returns the headers.
// Input may be http.Header, map[string][]string, Header, or *Header.
func (h *HeaderRuleSet) Apply(ctx context.Context, input any) (http.Header, errors.ValidationError) {
	var headers http.Header
	switch v := input.(type) {
	case http.Header:
		headers = v
	case map[string][]string:
		headers = http.Header(v)
	case *Header:
		headers = headerToHTTP(v)
	case Header:
		headers = headerToHTTP(&v)
	default:
		return nil, ToJSONAPIErrors(errors.Errorf(errors.CodeType, ctx, "http.Header, map[string][]string, or jsonapi.Header", reflect.ValueOf(input).Kind().String()), SourceHeader)
	}
	if err := h.Evaluate(ctx, headers); err != nil {
		return nil, ToJSONAPIErrors(err, SourceHeader)
	}
	return headers, nil
}

// Required returns true if Content-Type is required (default).
func (h *HeaderRuleSet) Required() bool {
	return h.contentRequired
}

// String returns a stable name for debugging.
func (h *HeaderRuleSet) String() string {
	return "HeaderRuleSet"
}

// Replaces implements rules.Rule[http.Header].
func (h *HeaderRuleSet) Replaces(r rules.Rule[http.Header]) bool { return false }

// Any returns the rule set as rules.RuleSet[any].
func (h *HeaderRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[http.Header](h)
}

var _ rules.RuleSet[http.Header] = (*HeaderRuleSet)(nil)
