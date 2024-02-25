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

func toPositional(s string) Arg {
	return Arg{argPositional, s}
}

func noMatch() Arg {
	return Arg{kind: argUnmatched}
}

func toValue(s string) Arg {
	return Arg{argValue, s}
}

func Short(toMatch rune) Arg {
	return Arg{argShort, string(toMatch)}
}

func Long(toMatch string) Arg {
	return Arg{argLong, toMatch}
}

func (a Arg) String() string {
	return a.s
}
