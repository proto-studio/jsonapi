package jsonapi

import (
	"context"
	"regexp"
	"strings"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
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

type AllValuesList struct{}

func (*AllValuesList) Values() []string {
	return nil
}

func (*AllValuesList) Contains(field string) bool {
	return true
}

func (*AllValuesList) doNotExtend() {}

type NoneValuesList struct{}

func (*NoneValuesList) Values() []string {
	return []string{}
}

func (*NoneValuesList) Contains(field string) bool {
	return false
}

func (*NoneValuesList) doNotExtend() {}

type fieldListMap map[string]bool

func (fl fieldListMap) Values() []string {
	keys := make([]string, 0, len(fl))
	for key := range fl {
		keys = append(keys, key)
	}
	return keys
}

func (fl fieldListMap) Contains(field string) bool {
	v, ok := fl[field]
	return ok && v
}

func (fieldListMap) doNotExtend() {}

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

type QueryData struct {
	Sort             []SortParam `validate:"sort"`
	Fields           map[string]ValueList
	Filters          map[string]string
	Include          ValueList      `validate:"include"`
	ExtensionMembers map[string]any `json:"-"`
}

var queryValueRuleSet = rules.Slice[string]().WithItemRuleSet(rules.String()).WithMaxLen(1)

var fieldKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^fields\[[^\]]+\]$`), "")
var filterKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^filter\[[^\]]+\]$`), "")

// Filter is only allowed on index GET requests
var filterRuleSet = rules.String().WithRule(HTTPMethodRule[string]("GET", "HEAD")).WithRule(IndexRule[string]())

var fieldsRuleSet = rules.Interface[ValueList]().WithCast(func(ctx context.Context, value any) (ValueList, errors.ValidationErrorCollection) {
	// Fields is allowed on all methods except DELETE
	method := MethodFromContext(ctx)

	if method == "DELETE" {
		return nil, errors.Collection(
			errors.Errorf(errors.CodeForbidden, ctx, "Fields are not allowed on DELETE requests"),
		)
	}

	var strs []string
	verrs := queryValueRuleSet.Apply(ctx, value, &strs)

	if verrs != nil {
		return nil, verrs
	}

	splitStrs := strings.Split(strs[0], ",")

	return NewFieldList(splitStrs...), nil
})

var includeRuleSet = fieldsRuleSet

var sortRuleSet = rules.Interface[[]SortParam]().WithCast(func(ctx context.Context, value any) ([]SortParam, errors.ValidationErrorCollection) {

	// Sort is only allowed on index GET requests
	method := MethodFromContext(ctx)
	id := IdFromContext(ctx)

	if method != "" && (id != "" || (method != "GET" && method != "HEAD")) {
		return nil, errors.Collection(
			errors.Errorf(errors.CodeForbidden, ctx, "Sort is only allowed on index GET requests"),
		)
	}

	var strs []string
	verrs := queryValueRuleSet.Apply(ctx, value, &strs)

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

var QueryStringBaseRuleSet rules.RuleSet[QueryData] = rules.Struct[QueryData]().
	WithDynamicKey(fieldKeyRule, fieldsRuleSet.Any()).
	WithDynamicKey(filterKeyRule, filterRuleSet.Any()).
	WithKey("include", includeRuleSet.Any()).
	WithDynamicBucket(fieldKeyRule, "Fields").
	WithDynamicBucket(filterKeyRule, "Filters").
	WithDynamicBucket(extKeyRule, "ExtensionMembers").
	WithKey("sort", sortRuleSet.Any())
