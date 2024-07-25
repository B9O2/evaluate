package expression

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// CELEvaluate CEL评估与执行器
type CELEvaluate struct {
	reg                             *types.Registry
	container                       string
	baseEnvOptions, extraEnvOptions []cel.EnvOption
	registeredTypes                 map[string]*cel.Type
}

// AddEnvOptions 增加CEL环境选项
func (ce *CELEvaluate) AddEnvOptions(eos ...cel.EnvOption) {
	ce.extraEnvOptions = append(ce.extraEnvOptions, eos...)
}

func (ce *CELEvaluate) TypeTransfer(t reflect.Type) (*cel.Type, bool) {
	var ct *cel.Type
	ok := true
	switch t.String() {
	case "int", "int64":
		ct = types.IntType
	case "bool":
		ct = types.BoolType
	case "string":
		ct = types.StringType
	case "error":
		ct = types.ErrorType
	case "[]uint8":
		ct = types.BytesType
	case "interface {}":
		ct = types.AnyType
	case "map[string]string":
		ct = types.NewMapType(types.StringType, types.StringType)
	default:
		ct, ok = ce.registeredTypes[t.String()]
	}
	return ct, ok
}

func (ce *CELEvaluate) ValueTransfer(a any) ref.Val {
	return ce.reg.NativeToValue(a)
}

// DeclareVariable 向CEL环境增加变量的捷径方法
func (ce *CELEvaluate) DeclareVariable(name string, a any) (*cel.Type, error) {
	var ct *cel.Type
	var ok bool
	if a == nil {
		ce.AddEnvOptions(cel.Variable(name, types.NullType))
		return types.NullType, nil
	}
	ta := reflect.TypeOf(a)
	if ct, ok = ce.TypeTransfer(ta); !ok {
		ce.AddEnvOptions(cel.Types(a))
		parts := strings.Split(ta.String(), ".")
		n := parts[len(parts)-1]
		ct = cel.ObjectType(ce.container + "." + n)
		if ct == nil {
			return nil, errors.New("raev: unknown value type '" + ta.String() + "'")
		}
		err := ce.reg.RegisterType(ct)
		if err != nil {
			return nil, err
		}
		ce.registeredTypes[ta.String()] = ct
	}
	ce.AddEnvOptions(cel.Variable(name, ct))
	return ct, nil
}

// AddFunction 向CEL环境增加函数的捷径方法
func (ce *CELEvaluate) AddFunction(name string, opt decls.FunctionOpt) {
	ce.AddEnvOptions(cel.Function(name, opt))
}

// Compile 编译表达式
func (ce *CELEvaluate) Compile(expression string) (cel.Program, error) {
	opts := append(ce.baseEnvOptions, cel.CustomTypeAdapter(ce.reg), cel.CustomTypeProvider(ce.reg))
	env, err := cel.NewEnv(append(opts, ce.extraEnvOptions...)...)
	if err != nil {
		return nil, err
	}
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("compile error, %v", issues.Err())
	}
	return env.Program(ast)
}

// Eval 自动调用Compile编译表达式并执行
func (ce *CELEvaluate) Eval(expression string, args map[string]interface{}) (ref.Val, *cel.EvalDetails, error) {
	prg, err := ce.Compile(expression)
	if err != nil {
		return nil, nil, err
	}
	return prg.Eval(args)
}

func NewCELEvaluate(container string) *CELEvaluate {
	ce := &CELEvaluate{
		reg:       types.NewEmptyRegistry(),
		container: container,
		baseEnvOptions: []cel.EnvOption{
			cel.Container(container),
		},
		registeredTypes: map[string]*cel.Type{},
	}
	return ce
}
