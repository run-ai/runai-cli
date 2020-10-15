package ui

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

const (
	GroupPrefix = "▿ "
	FieldPrefix = "  "
)

type (
	KeyValuePairs interface {
		Render(w io.Writer, rows interface{}) KeyValuePairs
		Error() error
	}

	KeyValuePairsOpt struct {
		// set the default for the root struct (any root fields will be hidden by default if is true)
		HideAllByDefault bool
		// which field paths to show
		Show []string
		// which field paths to hide
		Hide []string
		// map format name into a function
		Formatts FormattersByName
	}

	keyValuePairsData struct {
		base      Field
		modelType reflect.Type
		opt       KeyValuePairsOpt
		err       error
	}

	Field struct {
		BaseField
		isGroup     bool
		groupTag    *GroupTag
		perantField *Field
		fields      []Field
	}
)

func CreateKeyValuePairs(model interface{}, opt KeyValuePairsOpt) KeyValuePairs {

	data := keyValuePairsData{
		base: Field{
			isGroup: true,
			fields:  []Field{},
		},
		modelType: reflect.TypeOf(model),
		opt:       opt,
	}

	isShowAllByDefault := true

	if opt.HideAllByDefault {
		isShowAllByDefault = false
	} else if opt.Show != nil {
		// if there is at least one filed on the root of the struct
		for _, path := range opt.Show {
			if !strings.Contains(path, ".") {
				isShowAllByDefault = false
				break
			}
		}
	}

	data.addFields(data.modelType, []string{}, &data.base, isShowAllByDefault)

	return &data
}

func (td *keyValuePairsData) addFields(modelType reflect.Type, path []string, perantField *Field, showByDefult bool) {
	fieldsCount := modelType.NumField()
	for i := 0; i < fieldsCount; i++ {
		td.addField(modelType.Field(i), path, perantField, showByDefult)
	}
}

func (td *keyValuePairsData) addField(fieldType reflect.StructField, path []string, perantField *Field, showByDefult bool) {
	// if need to hide the field
	absolutePathPathStr := strings.Join(append(getPerentPath(path, perantField), fieldType.Name), ".")
	if td.opt.Hide != nil {
		if contains(td.opt.Hide, absolutePathPathStr) {
			showByDefult = false
		}
	}
	if td.opt.Show != nil {
		if contains(td.opt.Show, absolutePathPathStr) {
			showByDefult = true
		}
	}
	if isStructGroup(fieldType) {
		td.addGroup(fieldType, path, perantField, showByDefult)
		return
	}
	if !showByDefult {
		return
	}
	baseField, err := toBaseField(fieldType, td.opt.Formatts, path)
	if err != nil {
		td.err = err
		return
	}
	perantField.fields = append(perantField.fields, Field{
		BaseField:   baseField,
		perantField: perantField,
	})
}

func (td *keyValuePairsData) addGroup(field reflect.StructField, path []string, perantField *Field, showByDefult bool) {
	groupTag := NewGroupTag(field.Tag.Get(groupTagName))
	groupPath := append(path, field.Name)
	var grandPerantFiled *Field

	if !groupTag.Flatten {
		baseField, err := toBaseField(field, td.opt.Formatts, path)
		if err != nil {
			td.err = err
			return
		}
		if len(groupTag.Name) > 0 {
			baseField.Title = groupTag.Name
		}
		grandPerantFiled = perantField
		perantField = &Field{
			BaseField:   baseField,
			isGroup:     true,
			groupTag:    &groupTag,
			perantField: grandPerantFiled,
			fields:      []Field{},
		}
		// reset the path
		groupPath = []string{}
	}

	td.addFields(UnwrapTypePtr(field.Type), groupPath, perantField, showByDefult)

	if grandPerantFiled != nil {
		grandPerantFiled.fields = append(grandPerantFiled.fields, *perantField)
	}
}

func (td *keyValuePairsData) Render(w io.Writer, row interface{}) KeyValuePairs {
	if td.err != nil {
		return td
	}

	err := renderPairs(w, reflect.ValueOf(row), td.base, row, 0)

	if err != nil {
		td.err = err
	}

	return td
}

func (td *keyValuePairsData) Error() error {
	return td.err
}

/// helpers

func getPerentPath(path []string, perentGroup *Field) []string {
	if perentGroup != nil {
		return getPerentPath(append(perentGroup.Path, path...), perentGroup.perantField)
	}
	return path
}

func renderPairs(w io.Writer, t reflect.Value, base Field, root interface{}, indentation int) error {
	var err error
	for _, c := range base.fields {
		ftp := getNesstedVal(t, append(c.Path, c.Key))

		if c.isGroup && ftp != nil {
			// print the group title
			fmt.Fprint(w, multiStr("  ", indentation)+GroupPrefix+c.Title+"\t\n\t\n")
			err = renderPairs(w, *ftp, c, root, indentation+1)
			if err != nil {
				return err
			}
			continue
		}
		var val string

		// if the value is not nil
		if ftp != nil {
			ft := *ftp
			if c.Formmater != nil {
				val, err = c.Formmater(ft.Interface(), root)
				if err != nil {
					return err
				}
			} else {
				val = StringifyValue(ft)
			}
		}

		// set default value if it is an empty
		if len(val) == 0 {
			val = c.Defult
		}

		// skip empty values
		if len(val) == 0 {
			continue
		}

		// print
		//   Key         ⊜ Value
		//                 
		indentationStr := multiStr("  ", indentation)
		Line(w, indentationStr + FieldPrefix + c.Title, indentationStr + "⊜ " + val+ "\n\t" + indentationStr )
	}
	return nil
}
