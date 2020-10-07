package ui

import (
	"bytes"
	"testing"
	"text/tabwriter"
)

type Address struct {
	x int32 `title: "MY_x"`
	y *float32
}

type Person struct {
	name       string `title:"Name" def:"--"`
	Age        int
	address    Address  `group:"Addres"`
	addressPtr *Address `group:"Addres2"`
}

type mockWritter struct {
	data string
}

func (mw *mockWritter) Write(b []byte) (n int, err error) {
	mw.data += string(b[:])
	return 0, nil
}

func TestTable(t *testing.T) {
	t.Run("General Case", func(t *testing.T) {

		b := new(bytes.Buffer)
		expected := 
`            ┌⦿ Addres          ┌⦿ Addres2    
Name   Age  ├⚬ x       y       ├⚬ x        y
────   ───  │  ─       ─       │  ─        ─
moshe  30   │  3       552.39  │  3        552.39
--     50   │  3       552.39  │  3        552.39
`
		w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)
		addr := Address{y: float32ptr(552.38956564), x: 3}
		data := []Person{
			{"moshe", 30, addr, &addr},
			{"", 50, addr, &addr},
		}

		err := CreateTable(Person{}, TableOpt{}).Render(w, data).Error()
		if err == nil {
			_ = w.Flush()
		}

		got := b.String()

		if got != expected {
			t.Errorf("Strings dont match expected:\n\n%s\n result: \n\n%s", expected, got)
		}
	})

}

func float32ptr(f float32) *float32 {
	return &f
}
