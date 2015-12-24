package pipes

import (
	"fmt"
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
	Map() map[string]interface{}
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
func (p *PipeArgument) Map() map[string]interface{}   { return nil }
func (p *PipeArgument) IsNull() bool {
	if "void" == p.getType() {
		return true
	}
	return false
}

type PipeEvaluationArgument struct {
	PipeArgument
	blocks IPipeEntry
	result interface{}
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

func (p *PipeSelectionArgument) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	return p
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

func (p *PipeStringArgument) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	return p
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

func (p *PipePush) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipePush) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	rt.Push(arg)
	return arg
}

type PipePop struct {
	PipeAction
}

func (p *PipePop) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	return rt.Pop()
}

func (p *PipePop) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

/**
    swap stack top and argument
**/
type PipeSwap struct {
	PipeAction
}

func (p *PipeSwap) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipeSwap) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	result := rt.Pop()
	rt.Push(arg)
	return result
}

/**
    apply goquery.Selection.First to argument
**/
type PipeFirst struct {
	PipeAction
}

func (p *PipeFirst) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipeFirst) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	var selarg IPipeArgument
	if arg.getType() == "selection" {
		selarg = arg
	} else {
		selarg = arg.Exec(rt, arg)
	}
	if selarg.getType() == "selection" {
		return NewSelectionArgument(selarg.Selection().First())
	}
	panic(fmt.Sprintf("error - wrong arg type: %v", arg))
}

type PipeStore struct {
	PipeAction
}

func (p *PipeStore) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipeStore) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	name := p.getStringArgument(0, rt, arg)
	rt.vars[name] = arg
	return arg
}

type PipeClone struct {
	PipeStore
}

func (p *PipeClone) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	name := p.getStringArgument(0, rt, arg)
	rt.vars[name] = NewSelectionArgument(arg.Selection().Clone())
	return arg
}

func (p *PipeClone) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

type PipeRestore struct {
	PipeStore
}

func (p *PipeRestore) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipeRestore) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	name := p.getStringArgument(0, rt, arg)
	result, _ := rt.getVariable(name)
	if sel, ok := result.(*PipeSelectionArgument); ok {
		return sel
	} else if str, ok := result.(*PipeStringArgument); ok {
		return str
	} else if imap, ok := result.(*PipeMapArgument); ok {
		return imap
	} else {
		panic(fmt.Sprintf("unknown variable type to restore: %+v", result))
	}
}

type PipeScore struct {
	PipeAction
	score int
}

func (p *PipeScore) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipeScore) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	param := p.getStringArgument(0, rt, arg)
	if arg.getType() == "selection" {
		sel := arg.Selection()
		setAttr(sel, "score", param)
	}
	return arg
}

type PipeRemove struct {
	PipeAction
}

func (p *PipeRemove) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	if arg.getType() == "selection" {
		sel := arg.Selection()
		return NewSelectionArgument(sel.Remove())
	}
	return arg
}

func (p *PipeRemove) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

type PipeText struct {
	PipeAction
}

func (p *PipeText) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipeText) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	var selarg IPipeArgument
	if arg.getType() == "selection" {
		selarg = arg
	} else {
		selarg = arg.Exec(rt, arg)
	}
	if selarg.getType() == "selection" {
		return NewStringArgument(selarg.Selection().Text())
	}
	panic(fmt.Sprintf("wrong argument type: %v", arg))
}

type PipeAttr struct {
	PipeAction
	name  string
	value string
}

func (p *PipeAttr) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipeAttr) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {
	name := p.getStringArgument(0, rt, arg)
	if result, ok := arg.Selection().Attr(name); ok {
		return NewStringArgument(result)
	}
	return NewStringArgument("")
}

type PipeUnhumanDate struct {
	PipeAction
}

func (p *PipeUnhumanDate) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

type PipeAs struct {
	PipeAction
	name  string
	value interface{}
}

func (p *PipeAs) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

func (p *PipeAs) Exec(rt *PipeRuntime, arg IPipeArgument) IPipeArgument {

	value := arg.Exec(rt, arg)
	name := p.getStringArgument(0, rt, arg)

	if value.getType() == "string" {
		rt.vars[name] = value.String()
	} else if value.getType() == "selection" {
		rt.vars[name] = value.Selection()
	} else {
		rt.vars[name] = value
	}
	return arg
}

type PipeFrom struct {
	PipeAction
	name  string
	value interface{}
}

func (p *PipeFrom) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}

type PipeConcat struct {
	PipeAction
}

func (p *PipeConcat) Compile(exp *cExp) IPipeEntry {
	p.PipeAction.Compile(exp)
	return p
}
