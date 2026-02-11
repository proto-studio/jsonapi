package jsonapi

import (
	"regexp"

	"proto.zip/studio/validate/pkg/rules"
)

var atMembersKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^@`), "")

// Extension member names must be prefixed with namespace followed by colon (e.g., "version:id")
// Per spec, namespace must contain only a-z, A-Z, 0-9
var extKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+:.+`), "")
