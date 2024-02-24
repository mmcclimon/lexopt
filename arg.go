package lexopt

type argType int

const (
	argUnmatched argType = iota
	argShort
	argLong
)

type Arg struct {
	kind argType
	s    string
}

func longArg(s string) Arg {
	return Arg{argLong, s}
}

func noMatch() Arg {
	return Arg{kind: argUnmatched}
}

func (a Arg) String() string {
	return a.s
}
