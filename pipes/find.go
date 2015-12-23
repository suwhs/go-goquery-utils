package pipes

import (
	"fmt"
)

type PipeFind struct {
	PipeAction
}

func (p *PipeFind) Exec(r *PipeRuntime, arg IPipeArgument) IPipeArgument {
	fmt.Printf("PIPEFIND.EXEC:%v\n", p.args)
	if arg.getType() == "selection" {
		argstr := p.getStringArgument(0)
		fmt.Printf("\tgetStringArgument(0)=='%s'", argstr)
		sel := arg.Selection()
		res := sel.Find(argstr)
		return NewSelectionArgument(res)
	}
	panic("find must be executed with selection scope")
}

func (p *PipeFind) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}
