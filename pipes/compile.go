package pipes

import (
	"fmt"
	"strings"
	"text/scanner"
)

type cExp struct {
	token rune
	text  string
	exps  []cExp
}

func (exp *cExp) Compile() IPipeEntry {
	if exp.token == -2 {
		return exp.CompileCall()
	} else if exp.token == -6 {
		return NewStringArgument(exp.text)
	} else if exp.token == 0 {
		r := &PipeExpression{}
		r.Compile(exp)
		return r
	}
	fmt.Printf("'%v'.Compile(token:%d)\n", exp, exp.token)
	return nil
	// if exp.token == scanner.String {
	// 	if !strings.HasPrefix(exp.text, "\"") {
	// 		result := compilerFactory(exp)
	// 		if result == nil {
	// 			panic(fmt.Sprintf("exp '%v' returns nil", exp))
	// 		}
	// 	}
	// 	//		fmt.Printf("exp token: '%s'\n", exp.text)
	// 	arg := NewStringArgument(exp.text)
	// 	if entry, ok := arg.(IPipeEntry); ok {
	// 		return entry
	// 	}
	// 	if len(exp.exps) < 0 {
	// 		panic("string arg with args!")
	// 	}
	// } else if exp.text != "" {
	// 	return compilerFactory(exp)
	// } else if len(exp.exps) > 0 {
	// 	var chain []IPipeEntry
	// 	for _, v := range exp.exps {
	// 		// fmt.Printf("exp entry: %+v\n", exp.exps[i])
	// 		r := v.Compile()
	// 		fmt.Printf("compiled exp: %v", r)
	// 		chain = append(chain, r)
	// 	}
	// 	return &PipeExpression{chain: chain}
	// 	// result := &PipeExpression{}
	// 	// return result.Compile(exp)
	// }

	// return nil
}

func (exp *cExp) CompileCall() IPipeEntry {
	r := compilerFactory(exp)
	return r
}

func parsePipe(pipe string) []cExp {
	var s scanner.Scanner
	s.Init(strings.NewReader(pipe))
	return parse(&s)
}

/**
    operator(arguments) { block } | operator | (expression) | operator (arguments)

    |                             |          |              |          |         |
            |                     |          |              |          |         |
            |           |         |

      cExp       cExp[]    cExp      cExp          cExp         cExp       cExp
       \--------------------/
                 cExp                cExp          cExp         cExp       cExp


**/

func parse(s *scanner.Scanner) []cExp {
	/* */
	var tok rune
	var current = cExp{}
	var result = make([]cExp, 0, 10)
	for tok = s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		text := s.TokenText()
		if text == "(" || text == "{" {
			if current.exps == nil {
				current.exps = make([]cExp, 0, 10)
			}
			current.exps = append(current.exps, cExp{text: text, exps: parse(s)})
			if text == "{" && len(current.exps) == 1 {
				current.exps = current.exps[0].exps
			}
		} else if text == ")" || text == "}" {
			result = append(result, current)
			return result
		} else if text == "|" || text == "," {
			if current.text == "" {
				current.text = text
			}
			result = append(result, current)
			text = ""
			current = cExp{}
		} else {
			current.text = text
			current.token = tok
			// result = append(result, current)
			// current = cExp{}
		}

	}
	return append(result, current)
}

func compilePipe(pipe string) []IPipeEntry {
	exp := parsePipe(pipe)
	result := make([]IPipeEntry, len(exp))
	for i := range exp {
		result[i] = exp[i].Compile()
	}
	return result
}

func compilerFactory(exp *cExp) IPipeEntry {
	//	fmt.Printf("COMPILE_EXP: %+v", exp)
	var pe IPipeEntry
	text := exp.text
	if "find" == text {
		pe = &PipeFind{}
	} else if "text" == text {
		pe = &PipeText{}
	} else if "first" == text {
		pe = &PipeFirst{}
	} else if "struct" == text {
		pe = &PipeStruct{}
	} else if "as" == text {
		pe = &PipeAs{}
	} else if "unhumanPublishDate" == text {
		pe = &PipeUnhumanDate{}
	} else if "remove" == text {
		pe = &PipeRemove{}
	} else if "attr" == text {
		pe = &PipeAttr{}
	} else if "store" == text {
		pe = &PipeStore{}
	} else if "clone" == text {
		pe = &PipeClone{}
	} else if "restore" == text {
		pe = &PipeRestore{}
	} else if "push" == text {
		pe = &PipePush{}
	} else if "pop" == text {
		pe = &PipePop{}
	} else if "concat" == text {
		pe = &PipeConcat{}
	} else if "(" == text {
		pe = &PipeExpression{}
	}
	if pe != nil {
		fmt.Printf("%v -> call %v.Compile()\n", exp, pe)
		return pe.Compile(exp)
	}
	fmt.Printf("return nil for exp %+v\n", exp)
	return nil
}
