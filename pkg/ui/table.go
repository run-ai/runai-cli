package ui

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// DECLERATION: currently sub grouping are not supported

const (

	groupSeperation =  "| "

	// tag names
	titleTagName  = "title"
	defultTagName = "def"
	formatTagName = "format"
	groupTagName  = "group"

	// group flags
	// todo: flattenGroupFlag	= "flatten"
	// group features
	// todo: prefix = "prefix"
	// todo: groupMetaKey = "meta" // specify tag data for an struct

	// table feature

)

type FormatFunction = func(value interface{}, model interface{}) (string, error)
type FormatterMap = map[string]FormatFunction

type Column struct {
	Formmater                 FormatFunction
	Path 					  []string
	Key, Group, Title, Defult string
}

type tableData struct {
	columns   []Column
	modelType reflect.Type
	groups map[string]GroupTag
	opt 	  TableOpt
	err       error
}

type Table interface {
	Render(w io.Writer, rows interface{}) Table
	RenderHeader(w io.Writer) Table
	RenderRows(w io.Writer, rows interface{}) Table
	Error() error
}


func NewTag(tag string) Tag {
	t :=  Tag {
		Flags: map[string]bool{},
		Keys: map[string]string{},
	}
	tag = strings.TrimSpace(tag)
	tagSegments := strings.Split(tag, ",")
	for i,s := range tagSegments {
		if (i == 0) {
			t.Val = s
			continue
		}
		sub := strings.SplitN(tag, "=", 1)
		if len(sub) == 2 {
			t.Keys[sub[0]] = sub[1]
		} else {
			t.Flags[sub[0]] = true
		}
	}

	return t;

}

// Tag is a general tag structure
type Tag struct {
	Val string // the first value 
	Flags map[string]bool
	Keys map[string]string
}

type GroupTag struct {
	Name string
	// flags
	Flatten bool
	// keys
	Prefix string
}

type TableOpt struct {
	Hidden         []string
	CustomFormatts FormatterMap
}

func CreateTable(model interface{}, opt TableOpt) Table {
	columns := []Column{}

	
	td := tableData {
		columns: columns,
		modelType: reflect.TypeOf(model),
		opt: opt,
		groups: map[string]GroupTag{},
	}

	td.addFields(td.modelType, []string{}, NewGroupTag(""))

	return &td
}

func (td *tableData) addFields(modelType reflect.Type, path []string, groupTag GroupTag) {
	fieldsCount := modelType.NumField()
	for i := 0; i < fieldsCount; i++ {
		td.addField( modelType.Field(i), path, groupTag)
	}
}

func (td *tableData) addField(fieldType reflect.StructField, path []string, groupTag GroupTag) {
	if (isStructGroup(fieldType)) {
		td.addGroup(fieldType, path, groupTag)
		return
	} 
	column, err := toColumn(fieldType, td.opt.CustomFormatts, path, groupTag)
	if err != nil {
		td.err = err
		return
	}
	td.columns = append(td.columns, column)
}

func (td *tableData) addGroup(field reflect.StructField, path []string, groupTag GroupTag) {
	groupTagStr := field.Tag.Get(groupTagName)
	groupPath := append(path, field.Name)
	if len(groupTagStr) > 0 {
		groupTag = NewGroupTag(groupTagStr)
		groupPathStr := strings.Join(groupPath, ".")
		td.groups[groupPathStr] = groupTag
	}

	td.addFields(UnwrapTypePtr(field.Type), groupPath, groupTag)
}

func (td *tableData) Render(w io.Writer, rows interface{}) Table {
	return td.RenderHeader(w).RenderRows(w, rows)
}

func (td *tableData) RenderHeader(w io.Writer) Table {
	if td.err != nil {
		return td
	}


	// print the groups
	if len(td.groups) > 0 {
		groupsCount := map[string]int{};
		groups := []string{}
		for _, c := range td.columns {
			groupsCount[c.Group] = groupsCount[c.Group] + 1
		}
		for groupName, spaces := range groupsCount {
			if spaces == 0 {
				continue
			}
			tabs := make([]string, spaces)
			for i := range tabs{
				tabs[i]="\t"
			}
			if len(groupName) > 0 {
				groupName = groupSeperation + groupName
			}
			groups = append(groups, groupName + strings.Join(tabs, ""))
		}
		if len(groups) > 0 {
			fmt.Fprintln(w, strings.Join(groups, ""))
		}
	}

	titles := make([]string, len(td.columns))
	previousGroup := ""
	// todo add | before group
	for i, c := range td.columns {
		title := c.Title
		if i > 0 && previousGroup != c.Group {
			title = groupSeperation + title
		}
		previousGroup = c.Group
		titles[i] = title
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
		previousGroup := ""
		for i, c := range td.columns {
			var val string

			ftp := getNesstedVal(t, append(c.Path, c.Key))
			// if the value is not nil
			if ftp != nil {
				ft := *ftp
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

			if i > 0 && previousGroup != c.Group {
				val = groupSeperation + val
			}
			previousGroup = c.Group

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

//// helpers

func isStructGroup(field reflect.StructField) bool {
	isStruct := UnwrapTypePtr(field.Type).Kind() == reflect.Struct;

	group := field.Tag.Get(groupTagName)
	format := field.Tag.Get(formatTagName)

	return isStruct && len(group) > 0 && len(format) == 0
}

func NewGroupTag(tagStr string) GroupTag{
	tag := NewTag(tagStr)
	return GroupTag {
		Name: tag.Val,
		// todo: Flatten: tag.Flags[flattenGroupFlag] || len(tag.Val)==0,
	}
}

func toColumn(field reflect.StructField, formatMap FormatterMap, path []string, groupTag GroupTag) (Column, error) {
	var formaterFunc FormatFunction
	key := field.Name
	title := field.Tag.Get(titleTagName)
	def := field.Tag.Get(defultTagName)
	format := field.Tag.Get(formatTagName)

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
		Group:     groupTag.Name,
		Key:       key,
		Path:      path,
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

// UnwrapValuePtr recursively unwrap pointers
func UnwrapValuePtr(ft reflect.Value) *reflect.Value {
	if ft.Kind() == reflect.Ptr {
		if ft.IsNil() {
			return nil
		}
		return UnwrapValuePtr(ft.Elem())
	}
	return &ft
}

// UnwrapTypePtr recursively unwrap pointers
func UnwrapTypePtr(ft reflect.Type) reflect.Type {
	if ft.Kind() == reflect.Ptr {
		return UnwrapTypePtr(ft.Elem())
	}
	return ft
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


func getNesstedVal(t reflect.Value, path []string) (val *reflect.Value) {
	val = &t;

	for _, p := range path {
		// unwrap the pointers
		val = UnwrapValuePtr(val.FieldByName(p))
		if val == nil {
			return
		}
	}
	return 
}
