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

func SubTitle(w io.Writer, title string) {
	fmt.Fprintf(w, "\n%s: \n\n", title)
}

func Line(w io.Writer, fields ...string) {
	fmt.Fprintln(w, strings.Join(fields, "\t"))
}

func End(w io.Writer) {
	fmt.Fprintf(w, "\n")
}


func Bold(text interface{}) string {
	return fmt.Sprint("\033[1m",text,"\033[0m")
}
