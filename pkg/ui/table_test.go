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
	address    Address  `group:"Address"`
	addressPtr *Address `group:"Address 2"`
}

func TestTable(t *testing.T) {
	t.Run("General Case", func(t *testing.T) {
		expectedPath := "test_expected/table_test_1.txt"
		expected := from_file(expectedPath)
		b := new(bytes.Buffer)
		w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)
		addr := Address{y: float32ptr(552.38956564), x: 3}
		data := []Person{
			{"moshe", 30, addr, &addr},
			{"", 50, addr, &addr},
		}

		err := CreateTable(Person{}, TableOpt{}).Render(w, data).Error()
		if err == nil {
			_ = w.Flush()
		} else {
			t.Errorf("Failed to build the table, %s", err)
		}

		got := b.String()

		// for test debugging
		// record_at_file(expectedPath, got)

		if got != expected {
			t.Errorf("Strings dont match expected:\n\n%s\n result: \n\n%s", expected, got)
		}
	})

}

func float32ptr(f float32) *float32 {
	return &f
}
