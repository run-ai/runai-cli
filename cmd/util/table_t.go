package util

import (
	"os"
	"text/tabwriter"
)

type Person struct {
	name string `title:"Name" def:"--"`
	Age int
	address *float32 `title:"Addres"`
}

func RunTableExampleT() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	data := []Person{
		{"moshe", 30, float32ptr(552.38956564)},
		{"", 50, float32ptr(552.38956564)},
	}

	err := CreateTable(Person{}, TableOpt{}).Render(w, data).Error()
	if err == nil {
		_ = w.Flush()
	}

}

func float32ptr (f float32) *float32{
	return &f
}

