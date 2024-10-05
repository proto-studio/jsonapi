package jsonapi

import (
	"regexp"

	"proto.zip/studio/validate/pkg/rules"
)

var atMembersKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^@`), "")

var extKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^@`), "")
