package ui

import (
	"fmt"
	"reflect"
)

type BaseField struct {
	Formmater                 FormatFunction
	Path 					  []string
	Key, Title, Defult string
}

type Column struct {
	BaseField
	GroupID string
}

func toBaseField(field reflect.StructField, formatMap FormattersByName, path []string)  (BaseField, error) {
	var formaterFunc FormatFunction
	key := field.Name
	title := field.Tag.Get(titleTagName)
	def := field.Tag.Get(defultTagName)
	format := field.Tag.Get(formatTagName)

	if len(format) != 0 {
		f, found := formatMap[format]
		// if not found search in the default format
		if !found {
			f, found = DefaultFormatters[format]
		}

		if !found {
			return BaseField{}, fmt.Errorf("[Table] Not found format function for format name: %s  on field name: %s . Please make sure to include it in the TableOpt.CustomFormat", format, key)
		}
		formaterFunc = f
	}

	if len(title) == 0 {
		title = key
	}

	return BaseField{
		Title:     title,
		Defult:    def,
		Key:       key,
		Path:      path,
		Formmater: formaterFunc,
	}, nil
}


func toColumn(field reflect.StructField, formatMap FormattersByName, path []string, groupTag GroupTag) (Column, error) {
	baseField, err := toBaseField(field, formatMap, path )

	if err != nil {
		return Column{}, nil
	}
	return Column{
		BaseField: baseField,
		GroupID: groupTag.ID,
	}, nil
}