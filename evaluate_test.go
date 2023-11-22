package evaluate

import (
	"fmt"
	"reflect"
	"testing"
)

func Print(a string) string {
	fmt.Println(a)
	return "world!"
}

func TestNewEvaluate(t *testing.T) {
	e, err := NewEvaluate("expr")
	if err != nil {
		panic(err)
	}

	err = e.NewFunction("print", reflect.ValueOf(Print))
	if err != nil {
		panic(err)
		return
	}

	eval, err := e.Eval(`print("hello")`, map[string]any{})
	if err != nil {
		panic(err)
		return
	}
	fmt.Println(eval)

}
