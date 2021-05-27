package ui

import (
	"strconv"
	"strings"
)

type (

	// Tag is a general tag structure
	Tag struct {
		Val   string // the first value
		Flags map[string]bool
		Keys  map[string]string
	}

	GroupTag struct {
		Name string
		// to diff groups with the same name
		ID string
		// flags
		Flatten bool
		// keys
		Prefix string
	}
)

const (
	// tag names
	titleTagName   = "title"
	defaultTagName = "def"
	formatTagName  = "format"
	groupTagName   = "group"

	// group flags
	flattenGroupFlag = "flatten"
)

func NewTag(tag string) Tag {
	t := Tag{
		Flags: map[string]bool{},
		Keys:  map[string]string{},
	}
	tag = strings.TrimSpace(tag)
	tagSegments := strings.Split(tag, ",")
	for i, s := range tagSegments {
		if i == 0 {
			t.Val = s
			continue
		}
		sub := strings.SplitN(s, "=", 1)
		// check if it is a feature or a flag
		if len(sub) == 2 {
			t.Keys[sub[0]] = sub[1]
		} else {
			t.Flags[sub[0]] = true
		}
	}
	return t
}

func NewGroupTag(tagStr string) GroupTag {
	tag := NewTag(tagStr)
	groupID += 1
	return GroupTag{
		ID:      strconv.Itoa(groupID),
		Name:    tag.Val,
		Flatten: tag.Flags[flattenGroupFlag] || len(tag.Val) == 0,
	}
}
