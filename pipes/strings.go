package pipes

import "fmt"

func (p *PipeFind) String() string {
	return fmt.Sprintf("find(%v)", p.args)
}
