package evaluate

import (
	"errors"
	"fmt"
	"github.com/B9O2/evaluate/expression"
	"github.com/B9O2/raev"
	rTypes "github.com/B9O2/raev/types"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// Transfer 转换器的各个方法实现了转换的具体细节。
type Transfer struct {
	ce *expression.CELEvaluate
}

func (t *Transfer) methodTransfer(instance *cel.Type, m rTypes.Method) (decls.FunctionOpt, error) {
	var argTypes []*cel.Type
	var retType *cel.Type
	for _, param := range m.Params() {
		if at, ok := t.ce.TypeTransfer(param.Type); ok {
			argTypes = append(argTypes, at)
		} else {
			return nil, errors.New("Method Transfer: unknown type '" + param.String() + "'")
		}
	}
	for _, param := range m.ReturnParams() {
		if rt, ok := t.ce.TypeTransfer(param.Type); ok {
			retType = rt
		}
		break
	}
	var overload decls.FunctionOpt
	f := func(values ...ref.Val) ref.Val {
		var args []rTypes.ExtendObject
		for _, v := range values {
			args = append(args, v.(rTypes.ExtendObject))
		}
		rets, err := m.Call(args)
		if err != nil {
			return types.NewErr(err.Error())
		}
		if len(rets) > 0 {
			return rets[0].(ref.Val)
		}
		return nil
	}

	if instance != nil {
		overload = cel.MemberOverload(
			m.Name(),
			append([]*cel.Type{instance}, argTypes...),
			retType,
			cel.FunctionBinding(f),
		)
	} else {
		overload = cel.Overload(
			m.Name(),
			argTypes,
			retType,
			cel.FunctionBinding(f),
		)
	}
	return overload, nil
}

func (t *Transfer) ToClass(name string, c *rTypes.Class) (_ rTypes.ExtendClass, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	ct, err := t.ce.DeclareVariable(name, c.Raw().Interface())
	if err != nil {
		return nil, err
	}

	c.RangeMethods(func(method rTypes.Method) bool {
		var f decls.FunctionOpt
		f, err := t.methodTransfer(ct, method)
		if err != nil {
			return false
		}
		t.ce.AddFunction(CamelCaseToUnderscore(method.Name()), f)
		return true
	})

	return c.Raw().Interface(), nil
}

func (t *Transfer) ToMethod(method rTypes.Method) (rTypes.ExtendMethod, error) {
	return t.methodTransfer(nil, method)
}

func (t *Transfer) ToObject(a any) (rTypes.ExtendObject, error) {
	if c, ok := a.(*rTypes.Class); ok {
		return t.ce.ValueTransfer(c.Raw().Interface()), nil
	} else {
		return t.ce.ValueTransfer(a), nil
	}
}

func (t *Transfer) ToValue(obj rTypes.ExtendObject) (any, error) {
	o := obj.(ref.Val)
	switch name := o.Type().TypeName(); name {
	case "int":
		return o.Value().(int64), nil
	case "string":
		return o.Value().(string), nil
	case "bool":
		return o.Value().(bool), nil
	default:
		return o.Value(), nil
	}
}

func NewTransfer(ce *expression.CELEvaluate) *Transfer {
	return &Transfer{
		ce: ce,
	}
}

// Evaluate 表达式的评估执行器，支持结构体与函数自动映射
type Evaluate struct {
	ce *expression.CELEvaluate
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

func (e *Evaluate) NewClass(name string, source any, m any) (err error) {
	_, err = e.r.NewClass(name, source)
	if err != nil {
		return err
	}
	return e.NewFunction("new"+name, m)
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
	e := &Evaluate{
		ce: ce,
	}

	e.r = raev.NewRaev(
		decls.NewVariable("null", types.NewObjectType("expr.Null")),
		NewTransfer(ce),
	)

	return e, nil
}
