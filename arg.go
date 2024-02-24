package lexopt

type argType int

const (
	argUnmatched argType = iota
	argShort
	argLong
	argPositional
	argValue
)

type Arg struct {
	kind argType
	s    string
}

func toLong(s string) Arg {
	return Arg{argLong, s}
}

func toShort(b byte) Arg {
	return Arg{argShort, string(b)}
}

func toPositional(s string) Arg {
	return Arg{argPositional, s}
}

func noMatch() Arg {
	return Arg{kind: argUnmatched}
}

func toValue(s string) Arg {
	return Arg{argValue, s}
}

func (a Arg) String() string {
	return a.s
}
