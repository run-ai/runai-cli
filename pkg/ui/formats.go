package ui

import "fmt"

type (
	FormatFunction   = func(value interface{}, model interface{}) (string, error)
	FormattersByName = map[string]FormatFunction
)

var (
	DefaultFormatters = FormattersByName{
		"memory": BytesFormat,
		"%":      PrecantageFormat,
	}
)

/// formats

func BytesFormat(v interface{}, _ interface{}) (string, error) {
	switch t := v.(type) {
	case int64:
		return ByteCountIEC(t), nil
	case int:
		return ByteCountIEC(int64(t)), nil
	case float64:
		return ByteCountIEC(int64(t)), nil
	default:
		return fmt.Sprintf("%v", t), nil
	}
}

func PrecantageFormat(v interface{}, _ interface{}) (string, error) {

	return fmt.Sprintf("%.1f%%", v), nil
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
