package ui

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"text/tabwriter"
)



func check(e error) {
    if e != nil {
        panic(e)
    }
}

func TestKeyValuePairs(t *testing.T) {
	t.Run("General Case", func(t *testing.T) {

		b := new(bytes.Buffer)
		expected := from_file("expecteds/key_value_pairs_test_1.txt")

		w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)
		addr := Address{y: float32ptr(552.38956564), x: 3}
		data := Person{"moshe", 30, addr, &addr}
		// {"", 50, addr, &addr},
		err := CreateKeyValuePairs(Person{}, KeyValuePairsOpt{}).Render(w, data).Error()
		if err == nil {
			_ = w.Flush()
		} else {
			t.Errorf("Failed to build the table, %s", err)
		}

		got := b.String()


		if got != expected {
			fmt.Println(got)
			t.Errorf("Strings dont match expected:\n\n%s\nresult: \n\n%s", expected, got)
		}

	})

}


func from_file(fileName string) string {
	dat, err := ioutil.ReadFile(fileName)
	check(err)
    return string(dat)
}

// util to save the result at a file
func record_at_file(fileName string, content string) {
	f, err := os.Create("textoutput.txt")
	defer f.Close()
	check(err)
	f.WriteString(content)
}