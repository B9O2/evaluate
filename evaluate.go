package evaluate

import (
	"bytes"
	"errors"
	"reflect"
	"regexp"
	"strconv"

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
	ce          *expression.CELEvaluate
	t           *Transfer
	r           *raev.Raev
	templateReg *regexp.Regexp
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
	r, err := e.r.ObjectTransfer(ret, nil)
	if err != nil {
		return nil, err
	}
	return r.Interface(), nil
}

// EvalWithDefault 总是会返回运算结果，但在结果与默认值类型不同时将返回错误
func (e *Evaluate) EvalWithDefault(expr string, args map[string]any, defaultValue any) (any, error) {

	if len(expr) <= 0 {
		return defaultValue, nil
	}

	expectType := reflect.TypeOf(defaultValue)

	res, err := e.Eval(expr, args)
	if err != nil {
		return nil, err
	}
	value := res.(reflect.Value)
	if value.Type().String() == expectType.String() {
		return value.Interface(), nil
	} else {
		//部分类型间可以自动转换
		switch expectType.String() {
		case "string":
			switch v := value.Interface().(type) {
			case int64:
				return strconv.FormatInt(v, 10), nil
			case []uint8:
				return string(v), nil
			}

		case "[]uint8":
			switch v := value.Interface().(type) {
			case string:
				return []uint8(v), nil
			case int64:
				return []uint8(strconv.FormatInt(v, 10)), nil
			}
		}
		return value.Interface(), errors.New("the expression '" + string(expr) + "' should return " + expectType.String() + "<type>")
	}
}

// EvalWithTemplate 寻找模板表达式并执行
func (e *Evaluate) EvalWithTemplate(src []byte, args map[string]any) (_ []byte, err error) {
	res := e.templateReg.ReplaceAllFunc(src, func(s []byte) []byte {
		s = bytes.TrimLeft(s, "{")
		s = bytes.TrimRight(s, "}")
		result, err1 := e.EvalWithDefault(string(s), args, []uint8{})
		if err1 != nil {
			err = err1
			return []byte{}
		} else {
			return result.([]uint8)
		}
	})
	return res, err
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
	var err error
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

	e.templateReg, err = regexp.Compile(`\{\{(.+?)}}`)
	if err != nil {
		return nil, err
	}

	return e, nil
}
