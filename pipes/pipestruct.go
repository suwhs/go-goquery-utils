package pipes

import (
	"fmt"
)

type PipeStruct struct {
	PipeExpression
	expected []string
	fields   map[string]interface{}
}

func (p *PipeStruct) String() string {
	return fmt.Sprintf("PipeStruct(%+v [%+v])", p.fields, p.chain)
}

func (p *PipeStruct) Compile(exp *cExp) IPipeEntry {
	if len(exp.exps) != 2 {
		panic(fmt.Sprintf("struct () {} required arguments block and body block"))
	}
	var args []cExp
	args = exp.exps[0].exps
	blocks := &exp.exps[1].exps[0]
	p.expected = make([]string, len(args))
	for i := range args {
		p.expected[i] = unquote(args[i].text)
	}
	p.PipeExpression.Compile(blocks)
	return p
}

var debug bool = false

type PipeMapArgument struct {
	PipeArgument
	asMap map[string]interface{}
}

func NewPipeMapArgument(m map[string]interface{}) IPipeArgument {
	return &PipeMapArgument{asMap: m}
}
func (p *PipeMapArgument) getType() string             { return "map" }
func (p *PipeMapArgument) Map() map[string]interface{} { return p.asMap }
func (p *PipeMapArgument) String() string              { return fmt.Sprintf("<%v>", p.asMap) }
func (p *PipeMapArgument) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	return p
}

/**
run each chain from s
result of each chain must be as("<name>")

**/

func (p *PipeStruct) Exec(r *PipeRuntime, arg IPipeArgument) IPipeArgument {
	rt := r.Copy()
	var res map[string]interface{} = make(map[string]interface{})
	for _, v := range p.chain {
		v.Exec(rt, arg)
	}

	for _, v := range p.expected {
		if val, ok := rt.getVariable(v); !ok {
			panic(fmt.Sprintf("expected variable '%s', but not received\n", v))
		} else {
			res[v] = val
		}
	}
	result := &PipeMapArgument{}
	result.asMap = res
	return result
}
