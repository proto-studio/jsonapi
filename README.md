# Json:API

This library aims to provide the most complete Go implementation of the [Json:API spec](https://jsonapi.org/format/) possible.

It can be used for both Go clients and servers.

In addition to the normal marshalling/unmarshalling it also includes functionality for ensuring
that the incoming Json:API documents are valid according to the spec.

Features include:

- Works with `json.Unmarshal` and `json.Marshal`.
- Strictly typed attributes using Go generics.
- Support for nullable values.
- Powerful validation powered by [Proto Validate](https://github.com/proto-studio/protovalidate).
- Detailed error handling returns errors that are compliant with the Json:API spec.
- Different validation rules depending on the HTTP method.

The following features of the Json:API spec are currently supported:

- Request and response compliance checking.
- Sparse field sets.
- Filters.
- Cursor pagination.
- Sorting.
- Error objects.
- Version information.
- Meta information.
- Include / compound documents.
- Links (including nullable, string, and structured links).
- Relationships (including nullable relationships).
- Full error object support.

## Working with Nullable Values

An nil value is different in Json:API than a value that is not present.

For example, when patching a attribute if the value is explicitly set to null that indicates
the consumer wants the attribute to be removed. But if it is not present in the request that
means the consumer wants to leave the attribute as is (including if the attribute was already
at its default value).

The same is also true for relationships.

Likewise, in the server response a nil value for a relationship or link indicates that the
server has knowledge of the relationship/link and is explicitly indicating that it is not
present.

To handle these cases we use custom types for the fields that allow us to distinguish between
a value that is not present and a value that is present but set to nil. These types also provide
custom `MarshalJSON` and `UnmarshalJSON` functions to properly set the value to nil at the
appropriate times.

Likewise when unmarshaling with set the `Fields` value to a list of all the attributes that
were set in the Json request. This allows the caller to know what attributes were explicitly
set to an empty or nil value and which ones were just not present in the request.









