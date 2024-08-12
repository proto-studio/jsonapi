package jsonapi

import (
	"context"
	"regexp"
	"strings"

	"proto.zip/studio/validate"
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

type allFieldsList struct{}

func (*allFieldsList) Fields() []string {
	return nil
}

func (*allFieldsList) Contains(field string) bool {
	return true
}

func (*allFieldsList) doNotExtend() {}

type noneFieldsList struct{}

func (*noneFieldsList) Fields() []string {
	return []string{}
}

func (*noneFieldsList) Contains(field string) bool {
	return false
}

func (*noneFieldsList) doNotExtend() {}

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

type Include struct {
	Leaf string
	Path []string
}

type QueryData struct {
	Sort    []SortParam `validate:"sort"`
	Fields  map[string]FieldList
	Filters map[string]any
}

var queryValueRuleSet = validate.Array[string]().WithItemRuleSet(validate.String()).WithMaxLen(1)

var fieldKeyRule = validate.String().WithRegexp(regexp.MustCompile(`^fields\[[^\]]+\]$`), "")
var filterKeyRule = validate.String().WithRegexp(regexp.MustCompile(`^filter\[[^\]]+\]$`), "")

var filterRuleSet = validate.Interface[FieldList]().WithCast(func(ctx context.Context, value any) (FieldList, errors.ValidationErrorCollection) {
	strs, verrs := queryValueRuleSet.Run(ctx, value)

	if verrs != nil {
		return nil, verrs
	}

	itms := strings.Split(strs[0], ",")

	out := make(fieldListMap, len(itms))

	for _, itm := range itms {
		out[itm] = true
	}

	return out, nil
})

var sortRuleSet = validate.Interface[[]SortParam]().WithCast(func(ctx context.Context, value any) ([]SortParam, errors.ValidationErrorCollection) {
	strs, verrs := queryValueRuleSet.Run(ctx, value)

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

var QueryStringBaseRuleSet rules.RuleSet[QueryData] = validate.Object[QueryData]().
	WithDynamicKey(fieldKeyRule, filterRuleSet.Any()).
	WithDynamicBucket(fieldKeyRule, "Fields").
	WithDynamicBucket(filterKeyRule, "Filters").
	WithKey("sort", sortRuleSet.Any())
