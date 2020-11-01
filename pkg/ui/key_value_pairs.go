package ui

import (
	"fmt"
	"io"
	"reflect"
)

const (
	GroupPrefix = ""
	FieldPrefix = ""
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

func (kvd *keyValuePairsData) addFields(modelType reflect.Type, path []string, perantPair *PairMeta, showByDefult bool) {
	fieldsCount := modelType.NumField()
	for i := 0; i < fieldsCount; i++ {
		kvd.addField(modelType.Field(i), path, perantPair, showByDefult)
	}
}

func (kvd *keyValuePairsData) addField(fieldType reflect.StructField, path []string, perantPair *PairMeta, showByDefult bool) {
	
	absolutePath := getPerentPath( perantPair, append(path, fieldType.Name))

	showByDefult = kvd.opt.calcFieldShowByDefault(absolutePath, showByDefult )

	if isStructGroup(fieldType) {
		kvd.addGroup(fieldType, path, perantPair, showByDefult)
		return
	}
	if !showByDefult {
		return
	}
	fieldMeta, err := createFieldMeta(fieldType, kvd.opt.Formatts, path)
	if err != nil {
		kvd.err = err
		return
	}
	perantPair.children = append(perantPair.children, PairMeta{
		FieldMeta: fieldMeta,
		parent:    perantPair,
	})
}

func (kvd *keyValuePairsData) addGroup(field reflect.StructField, path []string, perantPair *PairMeta, showByDefult bool) {
	groupTag := NewGroupTag(field.Tag.Get(groupTagName))
	groupPath := append(path, field.Name)
	var grandPerantFiled *PairMeta

	if !groupTag.Flatten {
		fieldMeta, err := createFieldMeta(field, kvd.opt.Formatts, path)
		if err != nil {
			kvd.err = err
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

	kvd.addFields(UnwrapTypePtr(field.Type), groupPath, perantPair, showByDefult)

	if grandPerantFiled != nil {
		grandPerantFiled.children = append(grandPerantFiled.children, *perantPair)
	}
}

func (kvd *keyValuePairsData) Render(w io.Writer, row interface{}) KeyValuePairs {
	if kvd.err != nil {
		return kvd
	}

	err := renderPairChildren(w, reflect.ValueOf(row), kvd.base, row, 0)

	if err != nil {
		kvd.err = err
	}

	return kvd
}

func (kvd *keyValuePairsData) Error() error {
	return kvd.err
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

		// do nothing if it is a group but have no children
		if c.isGroup && len(c.children) == 0 {
			continue
		}

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