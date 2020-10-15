package ui

import (
	"fmt"
	"strings"
	"io"
)

// Title print a title
func Title(w io.Writer, title string) {
	fmt.Fprintf(w, "\n\n──────◆  %s  ◆──────\n\n", title)
}

func Line(w io.Writer, fields ...string) {
	fmt.Fprintln(w, strings.Join(fields, "\t"))
}

func End(w io.Writer) {
	fmt.Fprintf(w, "\n")
}
