package jsonapi_test

import (
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestMemberNameRule(t *testing.T) {
	rule := jsonapi.MemberNameRule{}

	valid := []string{"title", "createdAt", "author", "123field", "ext:version", "@id"}
	for _, name := range valid {
		testhelpers.MustEvaluate(t, rule, name)
	}

	invalid := []string{"field+name", "field,name", "field.name", "field[name", "field]name", "field!name", "field#name", "field name", "field@mid"}
	for _, name := range invalid {
		testhelpers.MustNotEvaluate(t, rule, name, errors.CodeUnexpected)
	}

	// Empty string returns CodeRequired
	testhelpers.MustNotEvaluate(t, rule, "", errors.CodeRequired)

	// @ only at start: allowed
	testhelpers.MustEvaluate(t, rule, "@member")
	// @ in middle: forbidden
	testhelpers.MustNotEvaluate(t, rule, "member@x", errors.CodeUnexpected)
}

func TestMemberNameRule_ReplacesAndString(t *testing.T) {
	rule := jsonapi.MemberNameRule{}
	if rule.Replaces(nil) {
		t.Error("MemberNameRule.Replaces should be false")
	}
	if s := rule.String(); s != "MemberNameRule" {
		t.Errorf("String(): got %q", s)
	}
}
