package evaluate

import (
	"errors"
	"fmt"
	"strings"

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
	*raev.BaseTransfer
	ce *expression.CELEvaluate
}

func (t *Transfer) MethodTransfer(instance *cel.Type, m rTypes.Method) (decls.FunctionOpt, error) {
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
			if rets[0] != nil {
				return rets[0].(ref.Val)
			}
		}
		return nil
	}

	if instance != nil {
		name := strings.ReplaceAll(instance.TypeName()+"_"+m.Name(), ".", "_")
		overload = cel.MemberOverload(
			name,
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
		f, err := t.MethodTransfer(ct, method)
		if err != nil {
			return false
		}
		t.ce.AddFunction(CamelCaseToUnderscore(method.Name()), f)
		return true
	})

	return c.Raw().Interface(), nil
}

func (t *Transfer) ToMethod(method rTypes.Method) (rTypes.ExtendMethod, error) {
	return t.MethodTransfer(nil, method)
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
	switch o.Type() {
	case types.IntType:
		return o.Value().(int64), nil
	case types.StringType:
		return o.Value().(string), nil
	case types.BoolType:
		return o.Value().(bool), nil
	case types.NullType:
		return nil, nil
	default:
		return o.Value(), nil
	}
}

func NewTransfer(ce *expression.CELEvaluate) *Transfer {
	return &Transfer{
		BaseTransfer: &raev.BaseTransfer{},
		ce:           ce,
	}
}
