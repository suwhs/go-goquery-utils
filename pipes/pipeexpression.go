package pipes

import (
	"fmt"
	"reflect"
)

func (p *PipeExpression) String() string {
	return fmt.Sprintf("%v", p.Blocks())
	// return fmt.Sprintf("PipeExpression (%+v)", p.chain)
}

func (p *PipeExpression) Compile(exp *cExp) IPipeEntry {
	// refactoring - acts all script as expression
	// update struct operator compiler as set of expressions
	/*

	*/
	p.chain = make([]IPipeEntry, 0, 5)
	for i := range exp.exps {
		pipe := exp.exps[i].Compile()
		fmt.Printf("\t compiled:%+v\n", reflect.TypeOf(pipe))
		p.chain = append(p.chain, pipe)
	}
	fmt.Printf("\nchain: %+v\nfrom:%+v\n", p.chain, exp)
	return p
}

func (p *PipeExpression) Exec(r *PipeRuntime, arg IPipeArgument) IPipeArgument {
	var result IPipeArgument = arg
	for i := range p.chain {
		result = Exec(p.chain[i], r, result)
	}
	return result
}

func (p *PipeExpression) Blocks() []IPipeEntry {
	return p.chain
}

/*
fmt.Printf("\ncompile action: '%s'\n", exp.text)
    if len(exp.exps>0) {

    }
*/
