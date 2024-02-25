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

func Short(toMatch string) Arg {
	if toMatch == "" {
		panic("argument to Short must not be empty")
	}

	if len(toMatch) > 1 {
		panic("argument to Short must be a single byte")
	}

	return toShort(toMatch[0])
}

func Long(toMatch string) Arg {
	if toMatch == "" {
		panic("argument to Long must not be empty")
	}

	return toLong(toMatch)
}

func (a Arg) String() string {
	return a.s
}
