package evaluate

import (
	"github.com/B9O2/raev/types"
	"unicode"
)

// CamelCaseToUnderscore 驼峰单词转下划线单词
func CamelCaseToUnderscore(s string) string {
	wordTrans := map[string]string{
		"string": "str",
	}
	var output []rune
	for i, r := range s {
		if i == 0 {
			output = append(output, unicode.ToLower(r))
		} else {
			if unicode.IsUpper(r) {
				output = append(output, '_')
			}

			output = append(output, unicode.ToLower(r))
		}
	}
	word := string(output)
	if to, ok := wordTrans[word]; ok {
		return to
	} else {
		return word
	}
}

func IgnoreInstanceParam(method *types.Method) {
	params := method.Params()
	if len(params) > 0 {
		method.SetParameters(params[1:])
	}
}
