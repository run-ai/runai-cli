package ui

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

const (
	titleTag  = "title"
	defultTag = "def"
	formatTag = "format"
	groupTag  = "group"
)

type FormatFunction = func(value interface{}, model interface{}) (string, error)
type FormatterMap = map[string]FormatFunction

type Column struct {
	Formmater                 FormatFunction
	Key, Group, Title, Defult string
}

type tableData struct {
	columns   []Column
	modelType reflect.Type
	err       error
}

type Table interface {
	Render(w io.Writer, rows interface{}) Table
	RenderHeader(w io.Writer) Table
	RenderRows(w io.Writer, rows interface{}) Table
	Error() error
}

type TableOpt struct {
	Hidden         []string
	CustomFormatts FormatterMap
}

func CreateTable(model interface{}, opt TableOpt) Table {
	columns := []Column{}

	t := reflect.TypeOf(model)

	fieldsCount := t.NumField()

	for i := 0; i < fieldsCount; i++ {
		column, err := toColumn(t.Field(i), opt.CustomFormatts)
		if err != nil {
			td := tableData{
				err: err,
			}
			return &td
		}
		columns = append(columns, column)
	}

	td := tableData{
		columns:   columns,
		modelType: t,
	}
	return &td
}

func (td *tableData) Render(w io.Writer, rows interface{}) Table {
	return td.RenderHeader(w).RenderRows(w, rows)
}

func (td *tableData) RenderHeader(w io.Writer) Table {
	if td.err != nil {
		return td
	}

	titles := make([]string, len(td.columns))

	// todo add | before group
	for i, c := range td.columns {
		titles[i] = c.Title
	}

	buffer := strings.Join(titles, "\t")
	fmt.Fprintln(w, buffer)
	return td
}

func (td *tableData) RenderRows(w io.Writer, rows interface{}) Table {
	if td.err != nil {
		return td
	}
	var err error
	data, err := interfaceToArrayOfInterface(rows)
	if err != nil {
		td.err = err
		return td
	}

	values := make([]string, len(td.columns))
	for _, r := range data {
		// todo add | before group
		t := reflect.ValueOf(r)
		for i, c := range td.columns {
			var val string
			ft := t.FieldByName(c.Key)

			// unwrap the pointers
			ftp := UnwrapPointers(ft)
			// if the value is not nil
			if ftp != nil {
				ft = *ftp
				if c.Formmater != nil {
					val, err = c.Formmater(ft.Interface(), r)
					if err != nil {
						td.err = err
						return td
					}
				} else {
					val = StringifyValue(ft)
				}
			}

			// set default value if it is an empty
			if len(val) == 0 {
				val = c.Defult
			}

			values[i] = val
		}

		buffer := strings.Join(values, "\t")
		fmt.Fprintln(w, buffer)

	}
	return td
}

func (td *tableData) Error() error {
	return td.err
}

func toColumn(field reflect.StructField, formatMap FormatterMap) (Column, error) {
	var key, title, def, group, format string
	var formaterFunc FormatFunction
	key = field.Name
	title = field.Tag.Get(titleTag)
	def = field.Tag.Get(defultTag)
	group = field.Tag.Get(groupTag)
	format = field.Tag.Get(formatTag)

	if len(format) != 0 {
		f, found := formatMap[format]
		if !found {
			return Column{}, fmt.Errorf("Not found format function to the format name: %s  on field %s . Please make sure to include it in the TableOpt.CustomFormat", format, key)
		}
		formaterFunc = f
	}

	if len(title) == 0 {
		title = key
	}

	return Column{
		Title:     title,
		Defult:    def,
		Group:     group,
		Key:       key,
		Formmater: formaterFunc,
	}, nil
}

func interfaceToArrayOfInterface(a interface{}) ([]interface{}, error) {
	object := reflect.ValueOf(a)
	items := make([]interface{}, object.Len())
	for i := 0; i < object.Len(); i++ {
		items[i] = object.Index(i).Interface()
	}
	return items, nil
}

func UnwrapPointers(ft reflect.Value) *reflect.Value {
	if ft.Kind() == reflect.Ptr {
		if ft.IsNil() {
			return nil
		}
		return UnwrapPointers(ft.Elem())
	}
	return &ft
}

// StringifyValue stringify any reflect.Value
func StringifyValue(ft reflect.Value) string {
	switch ft.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(ft.Int(), 10)
	case reflect.String:
		return ft.String()
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%16.2f", ft.Float())
	case reflect.Ptr:
		if ft.IsNil() {
			return ""
		}
		return StringifyValue(ft.Elem())
	default:
		// todo
		return string(ft.String())
	}
}
