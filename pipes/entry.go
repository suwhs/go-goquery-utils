package pipes

import (
	//	"fmt"
	"github.com/advancedlogic/goquery"
	"strings"
)

type PipeExpression struct {
	chain []IPipeEntry
}

/*
   find(find(\"div.article\")|attr(\"class\")|text)
   // here PipeEvaluationArgument
*/

type IPipeArgument interface {
	IPipeEntry
	getType() string
	Selection() *goquery.Selection
	Blocks() []IPipeEntry
	List() []IPipeArgument
	Int() int
	IsNull() bool
}

type PipeArgument struct {
	PipeExpression
}

func (p *PipeArgument) Compile(exp *cExp) IPipeEntry { return p }
func (p *PipeArgument) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	return p
}
func (p *PipeArgument) getType() string               { return "void" }
func (p *PipeArgument) String() string                { return "" }
func (p *PipeArgument) Selection() *goquery.Selection { return nil }
func (p *PipeArgument) Int() int                      { return 0 }
func (p *PipeArgument) Blocks() []IPipeEntry          { return nil }
func (p *PipeArgument) List() []IPipeArgument         { return nil }
func (p *PipeArgument) IsNull() bool {
	if "void" == p.getType() {
		return true
	}
	return false
}

type PipeEvaluationArgument struct {
	PipeArgument
	blocks IPipeEntry
}

func NewEvaluationArgument(eval IPipeEntry) IPipeArgument {
	return &PipeEvaluationArgument{blocks: eval}
}

func (p *PipeEvaluationArgument) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	return p.blocks.Exec(rt, arg)
}

type PipeSelectionArgument struct {
	PipeArgument
	arg *goquery.Selection
}

func (p *PipeSelectionArgument) getType() string               { return "selection" }
func (p *PipeSelectionArgument) Selection() *goquery.Selection { return p.arg }
func (p *PipeSelectionArgument) IsNull() bool                  { return p.arg == nil }

func NewSelectionArgument(sel *goquery.Selection) IPipeArgument {
	return &PipeSelectionArgument{arg: sel}
}

type PipeStringArgument struct {
	PipeArgument
	arg string
}

func (p *PipeStringArgument) getType() string { return "string" }
func (p *PipeStringArgument) String() string  { return p.arg }
func (p *PipeStringArgument) IsNull() bool    { return "" == p.arg }

func NewStringArgument(str string) IPipeArgument {
	if strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"") {
		return &PipeStringArgument{arg: str[1 : len(str)-1]}
	}
	return &PipeStringArgument{arg: str}
}

type PipeArgumentsList struct {
	PipeArgument
	args []IPipeArgument
}

func (p *PipeArgumentsList) getType() string       { return "list" }
func (p *PipeArgumentsList) List() []IPipeArgument { return p.args }

func NewPipeArgumentsList(args []IPipeArgument) IPipeArgument {
	return &PipeArgumentsList{
		args: args,
	}
}

/**
    operator and list of arguments
**/

type PipePush struct {
	PipeAction
}

func (p *PipePush) String() string { return "push" }

type PipePop struct {
	PipeAction
}

func (p *PipePop) String() string { return "pop" }

/**
    swap stack top and argument
**/
type PipeSwap struct {
	PipeAction
}

func (p *PipeSwap) String() string { return "swap" }

/**
    apply goquery.Selection.First to argument
**/
type PipeFirst struct {
	PipeAction
}

func (p *PipeFirst) String() string { return "first" }

type PipeStore struct {
	PipeAction
	name string
}

type PipeClone struct {
	PipeStore
}

type PipeRestore struct {
	PipeStore
}

type PipeScore struct {
	PipeAction
	score int
}

type PipeRemove struct {
	PipeAction
}

type PipeText struct {
	PipeAction
}

type PipeAttr struct {
	PipeAction
	name  string
	value string
}

type PipeUnhumanDate struct {
	PipeAction
}

type PipeAs struct {
	PipeAction
	name  string
	value interface{}
}

type PipeFrom struct {
	PipeAction
	name  string
	value interface{}
}

type PipeConcat struct {
	PipeExpression
}
