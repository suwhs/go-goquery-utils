package pipes

import (
	"fmt"
)

type PipeFind struct {
	PipeAction
}

func (p *PipeFind) Exec(r *PipeRuntime, arg IPipeArgument) IPipeArgument {
	if arg.getType() == "selection" {
		argstr := p.getStringArgument(0, r, arg)
		sel := arg.Selection()
		res := sel.Find(argstr)
		return NewSelectionArgument(res)
	}
	panic(fmt.Sprintf("find must be executed with selection scope, received: %+v", arg.getType()))
}

func (p *PipeFind) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}
