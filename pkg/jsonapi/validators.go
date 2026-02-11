package jsonapi

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Reserved member name characters per JSON:API 1.1 Section 5.8 (Document Member Names).
// Member names MUST NOT contain these except: ":" allowed in extension (namespace:member), "@" allowed at start for @-members.
var reservedMemberNameRunes = []rune{
	' ', '+', ',', '.', '[', ']', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '/', ';', '<', '=', '>', '?', '\\', '^', '`', '{', '|', '}', '~',
}

var reservedMemberNameTable map[rune]bool

// init builds the reservedMemberNameTable from reservedMemberNameRunes.
func init() {
	reservedMemberNameTable = make(map[rune]bool)
	for _, r := range reservedMemberNameRunes {
		reservedMemberNameTable[r] = true
	}
}

// MemberNameRule validates that a string is a valid JSON:API member name per the spec.
// Use with rules.String().WithRule(MemberNameRule) for dynamic keys, or Evaluate to check before adding members.
type MemberNameRule struct{}

// Evaluate implements rules.Rule[string].
func (MemberNameRule) Evaluate(ctx context.Context, value string) errors.ValidationError {
	if value == "" {
		return errors.Errorf(errors.CodeRequired, ctx, "member name required", "member name must not be empty")
	}
	for i, r := range value {
		if r == ':' {
			continue // allowed in extension member names (namespace:member)
		}
		if r == '@' {
			if i == 0 {
				continue // allowed at start for @-members
			}
			return errors.Errorf(errors.CodeUnexpected, ctx, "reserved character", "member name must not contain @ except at start (JSON:API @-members)")
		}
		if reservedMemberNameTable[r] {
			return errors.Errorf(errors.CodeUnexpected, ctx, "reserved character", "member name %q contains reserved character (JSON:API member names)", value)
		}
	}
	return nil
}

// Replaces implements rules.Rule[string].
func (MemberNameRule) Replaces(r rules.Rule[string]) bool { return false }

// String implements rules.Rule[string].
func (MemberNameRule) String() string { return "MemberNameRule" }

var _ rules.Rule[string] = MemberNameRule{}
