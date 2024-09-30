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

type FieldList interface {
	Fields() []string
	Contains(field string) bool
	doNotExtend()
}

type AllFieldsList struct{}

func (*AllFieldsList) Fields() []string {
	return nil
}

func (*AllFieldsList) Contains(field string) bool {
	return true
}

func (*AllFieldsList) doNotExtend() {}

type NoneFieldsList struct{}

func (*NoneFieldsList) Fields() []string {
	return []string{}
}

func (*NoneFieldsList) Contains(field string) bool {
	return false
}

func (*NoneFieldsList) doNotExtend() {}

type fieldListMap map[string]bool

func (fl fieldListMap) Fields() []string {
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

func NewFieldList(fields ...string) FieldList {
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
	Sort    []SortParam `validate:"sort"`
	Fields  map[string]FieldList
	Filters map[string]any
}

var queryValueRuleSet = rules.Slice[string]().WithItemRuleSet(rules.String()).WithMaxLen(1)

var fieldKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^fields\[[^\]]+\]$`), "")
var filterKeyRule = rules.String().WithRegexp(regexp.MustCompile(`^filter\[[^\]]+\]$`), "")

var filterRuleSet = rules.Interface[FieldList]().WithCast(func(ctx context.Context, value any) (FieldList, errors.ValidationErrorCollection) {
	var strs []string
	verrs := queryValueRuleSet.Apply(ctx, value, &strs)

	if verrs != nil {
		return nil, verrs
	}

	splitStrs := strings.Split(strs[0], ",")

	return NewFieldList(splitStrs...), nil
})

var sortRuleSet = rules.Interface[[]SortParam]().WithCast(func(ctx context.Context, value any) ([]SortParam, errors.ValidationErrorCollection) {
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
	WithDynamicKey(fieldKeyRule, filterRuleSet.Any()).
	WithDynamicKey(filterKeyRule, filterRuleSet.Any()).
	WithDynamicBucket(fieldKeyRule, "Fields").
	WithDynamicBucket(filterKeyRule, "Filters").
	WithKey("sort", sortRuleSet.Any())
