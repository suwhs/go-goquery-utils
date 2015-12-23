package pipes

import (
	"fmt"
	"github.com/advancedlogic/goquery"
	"log"
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
	blocks := &exp.exps[1]
	p.expected = make([]string, len(args))
	for i := range args {
		p.expected[i] = args[i].text
	}
	p.PipeExpression.Compile(blocks)
	return p
}

var debug bool = false

/**
run each chain from s
result of each chain must be as("<name>")

**/

func (p *PipeStruct) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	if debug || true {
		log.Printf("pipe chain: %+v", p.chain)
	}
	sr := NewPipeRuntime(s, true)
	/* p.fields = make(map[string]interface{})

	   for i := range p.chains {
	       log.Printf("exec struct chain %+v", p.chains[i])
	       var result interface{} = s
	       for k := range p.chains[i] {
	           result = Exec(p.chains[i][k], sr, result)
	       }

	       if name, ok := result.(*PipeAs); ok {
	           found := false
	           for l := range p.expected {
	               if p.expected[l] == name.name {
	                   found = true
	               }
	           }
	           if found {
	               p.fields[name.name] = sr.vars[name.name]
	           } else {
	               // unknown name for field
	               panic(fmt.Sprintf("unknown field '%s' for struct{%v}", name.name, p.fields))
	           }
	       } else {
	           // fatal error - chain must ends with struct's field name
	           panic(fmt.Sprintf("struct(...) chains must finished with `| as(\"name\")`, finished with %s from chain : %+v", result, p.chains[i]))
	       }
	       // copy expected fields from sr.vars to p.fields
	   } */
	if debug {
		log.Printf("return:'%+v'", sr)
	}
	return p
}
