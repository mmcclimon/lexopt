package lexopt

import "strings"

type Parser struct {
	Current Arg

	argv    []string
	idx     int
	state   state
	pending string
}

type state int

const (
	empty state = iota
	short
	pendingValue
	finished
)

func New(argv []string) *Parser {
	return &Parser{
		argv: argv,
		idx:  0,
	}
}

func (p *Parser) Next() bool {
	if p.idx >= len(p.argv) || p.state == finished {
		return false
	}

	defer func() { p.idx++ }()

	nextTok := p.argv[p.idx]
	switch {
	case nextTok == "--":
		p.state = finished
		return false

	case strings.HasPrefix(nextTok, "--"):
		p.state = empty

		nextTok = strings.TrimPrefix(nextTok, "--")
		before, after, hasEqual := strings.Cut(nextTok, "=")
		if hasEqual {
			p.pending = after
			p.state = pendingValue
		}
		p.Current = longArg(before)
		return true
	}

	// TODO
	// p.Current = p.argv[p.idx]
	return true
}

func (p *Parser) Err() error {
	return nil
}

func (p *Parser) Short(toMatch string) Arg {
	if toMatch == "" {
		panic("argument to Short must not be empty")
	}

	if p.state != short {
		return noMatch()
	}

	// TODO
	return noMatch()
}

func (p *Parser) Long(toMatch string) Arg {
	if toMatch == "" {
		panic("argument to Long must not be empty")
	}

	if p.state != empty && p.state != pendingValue {
		return noMatch()
	}

	return Arg{
		kind: argLong,
		s:    toMatch,
	}
}
