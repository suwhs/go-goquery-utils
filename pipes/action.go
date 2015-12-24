package pipes

import (
	"fmt"
)

type PipeAction struct {
	PipeExpression
	args []IPipeArgument
}

func (p *PipeAction) Compile(exp *cExp) IPipeEntry {
	if len(exp.exps) > 0 {
		// fmt.Printf("compile action: %s with args %v\n", exp.text, exp.exps[0])
		args := &exp.exps[0]
		p.args = p.CompileArgs(args)
		if len(exp.exps) > 1 {
			fmt.Printf("\t and body: %v", exp.exps[1:len(exp.exps)])
		}
	}
	// fmt.Printf("attached:%v\n", p.args)
	return p
}

func (p *PipeAction) CompileArgs(exp *cExp) []IPipeArgument {
	if exp.token != 0 {
		fmt.Printf("SYNTAX ERROR: args expression token must be 0")
		panic(":(")
	}
	result := make([]IPipeArgument, 0)
	for _, v := range exp.exps {
		if v.token == -6 {
			result = append(result, NewStringArgument(v.text))
		} else if v.token == -2 {
			result = append(result, NewEvaluationArgument(v.Compile()))
		}
	}
	return result
}
