package pipes

import (
	"fmt"
	"github.com/advancedlogic/goquery"
)

func (p *PipeAction) getArgument(i int, r *PipeRuntime, arg IPipeArgument) IPipeArgument {
	if i < len(p.args) {
		return p.args[i].Exec(r, arg)
	}
	panic("argument index out of bounds " + p.String())
}

func (p *PipeAction) getStringArgument(i int, r *PipeRuntime, iarg IPipeArgument) string {
	arg := p.getArgument(i, r, iarg)
	if arg != nil {
		if arg.getType() == "string" {
			return arg.String()
		} else {
			fmt.Printf("wrong argument type: '%v'\n", arg.getType())
		}
	}
	fmt.Printf("wrong argument type: %v", arg)
	panic(fmt.Sprintf("wrong argument type for %s expected: \"string\", received: %s", p.String(), p.getArgument(i, r, iarg)))
}

func (p *PipeAction) getSelectionArgument(i int, r *PipeRuntime, iarg IPipeArgument) *goquery.Selection {
	arg := p.getArgument(i, r, iarg)
	if arg != nil {
		if arg.getType() == "selection" {
			return arg.Selection()
		}
	}
	panic(fmt.Sprintf("wrong argument type for %s expected: \"selection\", received: %s", p.String(), p.getArgument(i, r, iarg)))
}
