package ui

import (
	"bytes"
	"fmt"
	"reflect"
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

type Temprature struct {
	value  int
	system string
}

type FridgeStatus struct {
	LastOpened int         `title:"Last Opened" format:"time"`
	Temprature *Temprature `title:"Temprature" format:"custom"`
}

type Fridge struct {
	Name   string        `title:"Name"`
	Status *FridgeStatus `group:"Status"`
}

func (tmpr Temprature) String() string {
	return fmt.Sprintf("%d (%v)", tmpr.value, tmpr.system)
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

	t.Run("User Provided Format", func(t *testing.T) {
		expectedPath := "test_expected/table_test_custom.txt"
		expected := from_file(expectedPath)
		b := new(bytes.Buffer)
		w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)
		temprature1 := Temprature{value: 4, system: "C"}
		temprature2 := Temprature{value: 40, system: "F"}
		data := []Fridge{
			{"vegtables", &FridgeStatus{LastOpened: 600, Temprature: &temprature1}},
			{"fruit", &FridgeStatus{LastOpened: 876, Temprature: &temprature2}},
		}

		err := CreateTable(Fridge{}, TableOpt{
			Formatts: map[string]FormatFunction{
				"custom": func(value, model interface{}) (string, error) {
					formatMethod := reflect.ValueOf(value).MethodByName("String")
					if formatMethod.IsValid() {
						result := formatMethod.Call([]reflect.Value{})
						return result[0].Interface().(string), nil
					}
					return "", fmt.Errorf("[UI::FieldMeta] no 'String' method found for type %s for 'custom' formatting", reflect.ValueOf(value).Type().Name())
				},
			},
		}).Render(w, data).Error()
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
