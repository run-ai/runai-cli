package ui

import (
	"fmt"
	"io"
)

// Title print a title
func Title(w io.Writer, title string) {
	fmt.Fprintf(w, "\n\n─────────────◼ %s ◼─────────────\n\n", title)
}


func End(w io.Writer) {
	fmt.Fprintf(w, "\n")
}
