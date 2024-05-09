package evaluate

import (
	"reflect"

	"github.com/B9O2/evaluate/expression"
	"github.com/B9O2/evaluate/middlewares"
	"github.com/B9O2/raev"
	rTypes "github.com/B9O2/raev/types"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/types"
)

// Evaluate 表达式的评估执行器，支持结构体与函数自动映射
type Evaluate struct {
	ce *expression.CELEvaluate
	t  *Transfer
	r  *raev.Raev
}

// DeclareVariable 声明变量
func (e *Evaluate) DeclareVariable(name string, a any) error {
	_, err := e.ce.DeclareVariable(name, a)
	return err
}

func (e *Evaluate) Eval(expr string, args map[string]any) (any, error) {
	ret, _, err := e.ce.Eval(expr, args)
	if err != nil {
		return nil, err
	}
	return e.r.ObjectTransfer(ret, nil)
}

func (e *Evaluate) NewClass(name string, source any, m any, extraMethods map[string]any) (err error) {

	var ms []rTypes.ClassMiddleware
	extraRawMethods := map[string]rTypes.Method{}
	if extraMethods != nil {
		for methodName, method := range extraMethods {
			rawMethod, err := e.r.NewRawMethod(methodName, reflect.ValueOf(method))
			if err != nil {
				return err
			}
			IgnoreInstanceParam(&rawMethod)
			extraRawMethods[methodName] = rawMethod
		}
		ms = append(ms, middlewares.NewExtraMethods(extraRawMethods))
	}

	_, err = e.r.NewClass(name, source, ms...)
	if err != nil {
		return err
	}

	if m != nil {
		return e.NewFunction("new"+name, m)
	}
	return nil
}

func (e *Evaluate) NewMemberFunction(celType *cel.Type, name string, m any) error {
	rawMethod, err := e.r.NewRawMethod(name, reflect.ValueOf(m))
	if err != nil {
		return err
	}
	IgnoreInstanceParam(&rawMethod)
	funcOpt, err := e.t.MethodTransfer(celType, rawMethod)
	if err != nil {
		return err
	}
	e.ce.AddFunction(name, funcOpt)
	return nil
}

func (e *Evaluate) NewFunction(name string, m any) error {
	overload, err := e.r.NewMethod(name, m)
	if err != nil {
		return err
	}
	e.ce.AddFunction(name, overload.(decls.FunctionOpt))
	return nil
}

func NewEvaluate(container string) (*Evaluate, error) {
	ce := expression.NewCELEvaluate(container)
	t := NewTransfer(ce)
	e := &Evaluate{
		ce: ce,
		t:  t,
	}

	e.r = raev.NewRaev(
		decls.NewVariable("null", types.NewObjectType("expr.Null")),
		t,
	)

	return e, nil
}
