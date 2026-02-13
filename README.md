# Json:API

[![Tests](https://github.com/proto-studio/jsonapi/actions/workflows/tests.yml/badge.svg)](https://github.com/proto-studio/jsonapi/actions/workflows/tests.yml)
[![GoDoc](https://pkg.go.dev/badge/proto.zip/studio/jsonapi)](https://pkg.go.dev/proto.zip/studio/jsonapi)
[![codecov](https://codecov.io/gh/proto-studio/jsonapi/graph/badge.svg)](https://codecov.io/gh/proto-studio/jsonapi)
[![Discord Chat](https://img.shields.io/badge/Discord-chat-blue?logo=Discord&logoColor=white)](https://proto.studio/social/discord)

This library is a Go implementation of the [Json:API v1.1 spec](https://jsonapi.org/format/) for both clients and servers.

Project goals:

1. Full marshalling and unmarshalling that works with `json.Marshal` and `json.Unmarshal`.
2. Spec-compliant validation with clear, actionable errors.
3. Type-safe resources and attributes using generics, with first-class support for nullable values and sparse fields.

Features:

- Works with `json.Unmarshal` and `json.Marshal`.
- Strictly typed attributes using Go generics.
- Support for nullable values and distinguishing “absent” from “present but null”.
- Validation powered by [ProtoValidate](https://github.com/proto-studio/protovalidate) (Go package: `proto.zip/studio/validate`).
- Errors that comply with the Json:API error object format.
- Validation rules that can vary by HTTP method.

Supported Json:API v1.1 features:

- Request and response compliance checking.
- Sparse field sets, filters, sorting.
- Error objects and version/meta information.
- Include and compound documents.
- Links (nullable, string, and structured) and relationships (including nullable).
- Extensions and @-members (captured and round-tripped).

You can use any extension or profile. The following profile is directly supported:

- [Cursor Pagination](https://jsonapi.org/profiles/ethanresnick/cursor-pagination/)

## Getting Started

### Quick Start

```bash
go get proto.zip/studio/jsonapi
```

Simple usage:

```go
package main

import (
	"context"
	"fmt"

	"proto.zip/studio/jsonapi"
	"proto.zip/studio/validate/pkg/rules"
)

func main() {
	attributesRuleSet := jsonapi.Attributes().
		WithJson().
		WithKey("title", rules.String().Any())

	ruleSet := jsonapi.NewSingleRuleSet[map[string]any]("articles", attributesRuleSet)
	ctx := context.Background()

	doc := `{"data":{"type":"articles","id":"1","attributes":{"title":"Hello"}}}`
	envelope, errs := ruleSet.Apply(ctx, doc)
	if errs != nil {
		fmt.Println(errs)
		return
	}
	fmt.Println(envelope.Data.Attributes["title"]) // Hello
}
```

## Working with Nullable Values

A nil value is different in Json:API than a value that is not present.

When patching an attribute, an explicit `null` means “remove this attribute.” Omitting the member means “leave it unchanged.” The same applies to relationships. In responses, a nil link or relationship means the server is explicitly indicating it is not present.

The library provides custom types with `MarshalJSON` and `UnmarshalJSON` so you can distinguish absent from null. When unmarshaling, `Datum.Fields` is set to the list of attribute names that appeared in the JSON, so you can tell which attributes were explicitly set (including to empty or null) and which were omitted.

## Versioning

This package follows conventional Go versioning. Any version before 1.0.0 is considered unstable and the API may change. Backwards incompatible changes in unstable releases will, when possible, deprecate first and be documented in release notes.

## Support

For community support, join the [ProtoStudio Discord Community](https://proto.studio/social/discord). For commercial support, contact [Curioso Industries](https://curiosoindustries.com).
