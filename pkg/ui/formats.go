package ui

import (
	"fmt"
	"strconv"
)

type (
	FormatFunction   = func(value interface{}, model interface{}) (string, error)
	FormattersByName = map[string]FormatFunction
)

var (
	DefaultFormatters = FormattersByName{
		"memory": BytesFormat,
		"%":      PrecantageFormat,
		"time":   TimeFormat,
	}
)

/// formats

func BytesFormat(v interface{}, _ interface{}) (string, error) {
	n, err := numericInterfaceToInt64(v)
	if err == nil {
		return ByteCountIEC(n), nil
	}
	
	return fmt.Sprintf("%v", v), err
}

func PrecantageFormat(v interface{}, _ interface{}) (string, error) {

	return fmt.Sprintf("%.1f%%", v), nil
}

func TimeFormat(v interface{}, _ interface{}) (string, error) {
	s, err := numericInterfaceToInt64(v)
	if err != nil {
		return "00:00:00", err
	}
	return fmt.Sprintf("%02d:%02d:%02d", int64(s / (60 * 60)) , int64(s / 60) % 60, s % 60), nil
}

/// utils

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}


func numericInterfaceToFloat64(n interface{}) (float64, error){
	var err error
	switch t := n.(type) {
	case int64:
		return float64(t), nil
	case int:
		return float64(t), nil
	case float32:
		return float64(t), nil
	case float64:
		return t, nil
	case string:
		asNum, convertErr := strconv.ParseFloat(t,64)
		if convertErr != nil {
			return asNum, nil
		}
		err = convertErr

	default: 
		err = fmt.Errorf("Unknown type")
	}
	return 0, err
}

func numericInterfaceToInt64(n interface{}) (int64, error){
	var err error
	switch t := n.(type) {
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case float32:
		return int64(t), nil
	case float64:
		return int64(t), nil
	case string:
		asNum, convertErr := strconv.ParseInt(t, 10, 64)
		if convertErr != nil {
			return asNum, nil
		}
		err = convertErr

	default: 
		err = fmt.Errorf("Unknown type")
	}
	return 0, err
}