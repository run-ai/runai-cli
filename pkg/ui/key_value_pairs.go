package ui

import (
	"fmt"
	"io"
	"reflect"
)

const (
	GroupPrefix = "▿ "
	FieldPrefix = "  "
	KeySuffix = ":"
	valuePrefix = ""
	IndentationPrefix = "  "
)

type (
	KeyValuePairs interface {
		Render(w io.Writer, rows interface{}) KeyValuePairs
		Error() error
	}

	KeyValuePairsOpt struct {
		DisplayOpt
		// map format name into a function
		Formatts FormattersByName
	}

	keyValuePairsData struct {
		base      PairMeta
		modelType reflect.Type
		opt       KeyValuePairsOpt
		err       error
	}

	PairMeta struct {
		FieldMeta
		parent   *PairMeta
		isGroup  bool
		groupTag *GroupTag
		children []PairMeta
	}
)

func CreateKeyValuePairs(model interface{}, opt KeyValuePairsOpt) KeyValuePairs {

	data := keyValuePairsData{
		base: PairMeta{
			isGroup:  true,
			children: []PairMeta{},
		},
		modelType: reflect.TypeOf(model),
		opt:       opt,
	}

	isShowAllByDefault := opt.rootShowByDefault()

	data.addFields(data.modelType, []string{}, &data.base, isShowAllByDefault)

	return &data
}

func (td *keyValuePairsData) addFields(modelType reflect.Type, path []string, perantPair *PairMeta, showByDefult bool) {
	fieldsCount := modelType.NumField()
	for i := 0; i < fieldsCount; i++ {
		td.addField(modelType.Field(i), path, perantPair, showByDefult)
	}
}

func (td *keyValuePairsData) addField(fieldType reflect.StructField, path []string, perantPair *PairMeta, showByDefult bool) {
	
	absolutePath := getPerentPath( perantPair, append(path, fieldType.Name))

	showByDefult = td.opt.calcFiledShowByDefult(absolutePath, showByDefult )

	if isStructGroup(fieldType) {
		td.addGroup(fieldType, path, perantPair, showByDefult)
		return
	}
	if !showByDefult {
		return
	}
	fieldMeta, err := createFieldMeta(fieldType, td.opt.Formatts, path)
	if err != nil {
		td.err = err
		return
	}
	perantPair.children = append(perantPair.children, PairMeta{
		FieldMeta: fieldMeta,
		parent:    perantPair,
	})
}

func (td *keyValuePairsData) addGroup(field reflect.StructField, path []string, perantPair *PairMeta, showByDefult bool) {
	groupTag := NewGroupTag(field.Tag.Get(groupTagName))
	groupPath := append(path, field.Name)
	var grandPerantFiled *PairMeta

	if !groupTag.Flatten {
		fieldMeta, err := createFieldMeta(field, td.opt.Formatts, path)
		if err != nil {
			td.err = err
			return
		}
		if len(groupTag.Name) > 0 {
			fieldMeta.Title = groupTag.Name
		}
		grandPerantFiled = perantPair
		perantPair = &PairMeta{
			FieldMeta: fieldMeta,
			isGroup:   true,
			groupTag:  &groupTag,
			parent:    grandPerantFiled,
			children:  []PairMeta{},
		}
		// reset the path
		groupPath = []string{}
	}

	td.addFields(UnwrapTypePtr(field.Type), groupPath, perantPair, showByDefult)

	if grandPerantFiled != nil {
		grandPerantFiled.children = append(grandPerantFiled.children, *perantPair)
	}
}

func (td *keyValuePairsData) Render(w io.Writer, row interface{}) KeyValuePairs {
	if td.err != nil {
		return td
	}

	err := renderPairChildren(w, reflect.ValueOf(row), td.base, row, 0)

	if err != nil {
		td.err = err
	}

	return td
}

func (td *keyValuePairsData) Error() error {
	return td.err
}

/// helpers

func getPerentPath(perentPairMeta *PairMeta, currentPath []string) []string {
	if perentPairMeta != nil {
		var perentGroupPath []string
		if len(perentPairMeta.Key) > 0 {
			perentGroupPath = append(perentPairMeta.Path, perentPairMeta.Key)
		} else {
			perentGroupPath = perentPairMeta.Path
		}
		return getPerentPath(
			perentPairMeta.parent,
			append(
				perentGroupPath,
				currentPath...
			), 
		)
	}
	return currentPath
}

func renderPairChildren(w io.Writer, t reflect.Value, pair PairMeta, root interface{}, indentation int) error {
	var err error
	for _, c := range pair.children {
		fieldTypeP := getNesstedVal(t, append(c.Path, c.Key))
		indentationStr := multiStr(IndentationPrefix, indentation)

		if c.isGroup && fieldTypeP != nil {
			// print the group title
			groupTitleOutput := indentationStr + GroupPrefix + c.Title + KeySuffix
			fmt.Fprint(w, groupTitleOutput+"\n")
			err = renderPairChildren(w, *fieldTypeP, c, root, indentation+1)
			if err != nil {
				return err
			}
			continue
		}
		var val string

		// if the value is not nil
		if fieldTypeP != nil {
			ft := *fieldTypeP
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

		// print:
		//   Key         ➯ Value
		keyOutput := indentationStr + FieldPrefix + c.Title + KeySuffix
		valueOutput :=  valuePrefix + val
		Line(w, keyOutput, valueOutput)
	}
	return nil
}