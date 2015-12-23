package pipes

import (
	"fmt"
	"github.com/advancedlogic/goquery"
	// "log"
	// "strconv"
	// "strings"
)

func (p *PipeAction) getArgument(i int) IPipeArgument {
	if i < len(p.args) {
		return p.args[i]
	}
	panic("argument index out of bounds " + p.String())
}

func (p *PipeAction) getStringArgument(i int) string {
	arg := p.getArgument(i)
	if arg != nil {
		if arg.getType() == "string" {
			fmt.Printf("return arg:%v", arg)
			return arg.String()
		} else {
			fmt.Printf("wrong argument type: %v", arg)
		}
	}
	fmt.Printf("wrong argyment type: %v", arg)
	panic(fmt.Sprintf("wrong argument type for %s expected: \"string\", received: %s", p.String(), p.getArgument(i)))
}

func (p *PipeAction) getSelectionArgument(i int) *goquery.Selection {
	arg := p.getArgument(i)
	if arg != nil {
		if arg.getType() == "selection" {
			return arg.Selection()
		}
	}
	panic(fmt.Sprintf("wrong argument type for %s expected: \"selection\", received: %s", p.String(), p.getArgument(i)))
}
