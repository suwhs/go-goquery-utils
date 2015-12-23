package pipes

type IPipeEntry interface {
	Compile(exp *cExp) IPipeEntry
	Exec(runtime *PipeRuntime, arg IPipeArgument) IPipeArgument
	String() string
}
