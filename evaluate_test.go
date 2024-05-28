package evaluate

import (
	"fmt"
	"testing"
)

// 必须具有返回值，可以返回多个值，最后返回值可以为error
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

	eval, err := e.Eval(`print("hello")`, map[string]any{
		"str": "hello world!",
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(eval)

}
