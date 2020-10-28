package ui

import (
	"fmt"
	"reflect"
)

type FieldMeta struct {
	Formmater                 FormatFunction
	Path 					  []string
	Key, Title, Defult string
}


func createFieldMeta(field reflect.StructField, formatMap FormattersByName, path []string)  (FieldMeta, error) {
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
			return FieldMeta{}, fmt.Errorf("[UI::FieldMeta] Not found format function for format name: %s  on field name: %s . Please make sure to include it in the TableOpt.CustomFormat", format, key)
		}
		formaterFunc = f
	}

	if len(title) == 0 {
		title = key
	}

	return FieldMeta{
		Title:     title,
		Defult:    def,
		Key:       key,
		Path:      path,
		Formmater: formaterFunc,
	}, nil
}
