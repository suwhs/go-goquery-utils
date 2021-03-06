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
	panic("unhandled compilation error")
	return nil
}

func (exp *cExp) CompileCall() IPipeEntry {
	r := compilerFactory(exp)
	return r
}

func parsePipe(pipe string) []cExp {
	var s scanner.Scanner
	s.Init(strings.NewReader(pipe))
	s.Error = func(s *scanner.Scanner, msg string) {

	}
	r := parse(&s)
	return r
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
		return pe.Compile(exp)
	}
	fmt.Printf("return nil for exp %+v\n", exp)
	return nil
}
