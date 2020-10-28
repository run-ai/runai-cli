package ui

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// DECLERATION: currently sub grouping are not supported

const (
	groupSeperation = "╭⦿ "
	titleSeperation = "├⚬ "
	rowsSeperation  = "│  "
)

var (
	groupId = 0
)

type (
	tableData struct {
		columns   []Column
		modelType reflect.Type
		groups    []GroupTag
		opt       TableOpt
		err       error
	}

	Table interface {
		Render(w io.Writer, rows interface{}) Table
		RenderHeader(w io.Writer) Table
		RenderRows(w io.Writer, rows interface{}) Table
		Error() error
	}

	TableOpt struct {
		DisplayOpt
		// map format name into a function
		Formatts FormattersByName
	}

	Column struct {
		FieldMeta
		GroupID string
	}
)

func CreateTable(model interface{}, opt TableOpt) Table {
	columns := []Column{}

	td := tableData{
		columns:   columns,
		modelType: reflect.TypeOf(model),
		opt:       opt,
	}

	isShowAllByDefault := opt.rootShowByDefault()

	defaultGroup := NewGroupTag("")
	td.groups = []GroupTag{defaultGroup}

	td.addFields(td.modelType, []string{}, defaultGroup, isShowAllByDefault)

	return &td
}

func (td *tableData) addFields(modelType reflect.Type, path []string, groupTag GroupTag, showByDefult bool) {
	fieldsCount := modelType.NumField()
	for i := 0; i < fieldsCount; i++ {
		td.addField(modelType.Field(i), path, groupTag, showByDefult)
	}
}

func (td *tableData) addField(fieldType reflect.StructField, path []string, groupTag GroupTag, showByDefult bool) {

	showByDefult = td.opt.calcFieldShowByDefault(append(path, fieldType.Name), showByDefult )
	
	if isStructGroup(fieldType) {
		td.addGroup(fieldType, path, groupTag, showByDefult)
		return
	}
	if !showByDefult {
		return
	}
	column, err := toColumn(fieldType, td.opt.Formatts, path, groupTag)
	if err != nil {
		td.err = err
		return
	}
	td.columns = append(td.columns, column)
}

func (td *tableData) addGroup(field reflect.StructField, path []string, groupTag GroupTag, showByDefult bool) {
	groupTag = NewGroupTag(field.Tag.Get(groupTagName))
	groupPath := append(path, field.Name)
	td.groups = append(td.groups, groupTag)

	td.addFields(UnwrapTypePtr(field.Type), groupPath, groupTag, showByDefult)
}

func (td *tableData) Render(w io.Writer, rows interface{}) Table {
	return td.RenderHeader(w).RenderRows(w, rows)
}

func (td *tableData) RenderHeader(w io.Writer) Table {
	if td.err != nil {
		return td
	}

	// add the groups
	if len(td.groups) > 1 {
		groupsCount := map[string]int{}
		groups := []string{}
		for _, c := range td.columns {
			groupsCount[c.GroupID] = groupsCount[c.GroupID] + 1
		}
		for i, tag := range td.groups {
			groupName := tag.Name
			if tag.Flatten {
				groupName = ""
			}
			spaces := groupsCount[tag.ID]
			if spaces == 0 {
				continue
			}
			tabs := make([]string, spaces)
			for i := range tabs {
				tabs[i] = "\t"
			}
			if i > 0 && !tag.Flatten {
				groupName = groupSeperation + groupName
			}
			groups = append(groups, groupName+strings.Join(tabs, ""))
			i++
		}
		if len(groups) > 0 {
			fmt.Fprintln(w, strings.Join(groups, ""))
		}
	}

	titles := make([]string, len(td.columns))
	titlesBottomBorder := make([]string, len(td.columns))

	previousGroup := "1"
	for i, c := range td.columns {
		title := c.Title
		border := multiStr("─", len(title))
		if i > 0 && previousGroup != c.GroupID {
			title = titleSeperation + title
			border = rowsSeperation + border
		}
		previousGroup = c.GroupID
		titles[i] = title
		titlesBottomBorder[i] = border
	}

	fmt.Fprintln(w, strings.Join(titles, "\t"))
	fmt.Fprintln(w, strings.Join(titlesBottomBorder, "\t"))

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

			if i > 0 && previousGroup != c.GroupID {
				val = rowsSeperation + val
			}
			previousGroup = c.GroupID

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

func toColumn(field reflect.StructField, formatMap FormattersByName, path []string, groupTag GroupTag) (Column, error) {
	fieldMeta, err := createFieldMeta(field, formatMap, path )

	if err != nil {
		return Column{}, err
	}
	return Column{
		FieldMeta: fieldMeta,
		GroupID: groupTag.ID,
	}, nil
}

func isStructGroup(field reflect.StructField) bool {
	isStruct := UnwrapTypePtr(field.Type).Kind() == reflect.Struct

	group := field.Tag.Get(groupTagName)
	format := field.Tag.Get(formatTagName)

	return isStruct && len(group) > 0 && len(format) == 0
}

