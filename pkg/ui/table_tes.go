package ui

import (
	"os"
	"text/tabwriter"
)


type Address struct {
	x int32 `title: "MY_x"`
	y *float32
}

type Person struct {
	name string `title:"Name" def:"--"`
	Age int
	address Address `group:"Addres"`
	addressPtr *Address `group:"Addres2"`
}

func RunTableExampleTes() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	addr := Address{y:float32ptr(552.38956564), x:3}
	data := []Person{
		{"moshe", 30, addr, &addr},
		{"", 50,  addr, &addr},
	}

	err := CreateTable(Person{}, TableOpt{}).Render(w, data).Error()
	if err == nil {
		_ = w.Flush()
	}

}

func float32ptr (f float32) *float32{
	return &f
}

