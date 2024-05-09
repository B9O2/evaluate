package evaluate

import (
	"fmt"
	"testing"
)

func Print(a string) string {
	fmt.Println(a)
	return "world!"
}

func SliceHere() []string {
	return []string{"ok"}
}

func TestNewEvaluate(t *testing.T) {
	e, err := NewEvaluate("expr")
	if err != nil {
		panic(err)
	}

	err = e.DeclareVariable("str", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = e.NewFunction("print", Print)
	if err != nil {
		panic(err)
	}

	err = e.NewFunction("slice", SliceHere)
	if err != nil {
		panic(err)
	}

	eval, err := e.Eval(`print()`, map[string]any{
		"str": "hello world!",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(eval)

}
