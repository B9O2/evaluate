package expression

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	"testing"
)

type Test struct {
	a string
}

func NewTest(a string) *Test {
	return &Test{a: a}
}

func (t Test) Hello() bool {
	fmt.Println(t.a)
	return true
}

func Print(a string) string {
	fmt.Println(a)
	return "flag"
}

func TestNewCELEvaluate(t *testing.T) {
	ce := NewCELEvaluate("expr")
	responseType, err := ce.DeclareVariable("response", (*HTTPResponseType)(nil))
	if err != nil {
		return
	}
	ce.AddFunction("newResponse", cel.Overload("new_response", nil, responseType, cel.FunctionBinding(func(values ...ref.Val) ref.Val {
		return ce.reg.NativeToValue(&HTTPResponseType{Status: 200})
	})))

	eval, _, err := ce.Eval(`newResponse().status`, map[string]any{
		"response": &HTTPResponseType{
			Url:         nil,
			Status:      200,
			Body:        nil,
			Headers:     nil,
			ContentType: "",
			Latency:     0,
			Raw:         nil,
			Title:       "",
			RawHeader:   nil,
			RawCert:     nil,
		},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(eval)
}
