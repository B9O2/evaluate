package middlewares

import "github.com/B9O2/raev/types"

type ExtraMethods map[string]types.Method

func (em ExtraMethods) Handle(class *types.Class) {
	for name, m := range em {
		class.SetMethod(name, m)
	}
}

func NewExtraMethods(methods map[string]types.Method) ExtraMethods {
	return methods
}
