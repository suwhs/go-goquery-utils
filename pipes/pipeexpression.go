package pipes

import (
	"fmt"
)

func (p *PipeExpression) String() string {
	return fmt.Sprintf("%v", p.chain)
}

func (p *PipeExpression) Compile(exp *cExp) IPipeEntry {
	p.chain = make([]IPipeEntry, 0, 5)
	for i := range exp.exps {
		pipe := exp.exps[i].Compile()
		p.chain = append(p.chain, pipe)
	}
	return p
}

func (p *PipeExpression) Exec(r *PipeRuntime, arg IPipeArgument) IPipeArgument {
	var result IPipeArgument = arg
	for _, v := range p.chain {
		result = Exec(v, r, result)
	}
	return result
}

func (p *PipeExpression) Blocks() []IPipeEntry {
	return p.chain
}
