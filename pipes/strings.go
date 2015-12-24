package pipes

import "fmt"

func (p *PipeFind) String() string {
	return fmt.Sprintf("find(%v)", p.args)
}

func (p *PipeAs) String() string {
	return fmt.Sprintf("as(%v)", p.args)
}

func (p *PipeText) String() string {
	return "text"
}

func (p *PipeAttr) String() string {
	return fmt.Sprintf("as(%v)", p.args)
}

func (p *PipeFirst) String() string {
	return "first"
}

func (p *PipePush) String() string { return "push" }

func (p *PipePop) String() string    { return "pop" }
func (p *PipeSwap) String() string   { return "swap" }
func (p *PipeRemove) String() string { return "remove" }
