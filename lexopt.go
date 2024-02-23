package lexopt

import "strings"

type Parser struct {
	Current string

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
		before, after, hasEqual := strings.Cut(nextTok, "=")
		if hasEqual {
			p.pending = after
			p.state = pendingValue
		}
		p.Current = strings.TrimPrefix(before, "--")
		return true
	}

	p.Current = p.argv[p.idx]
	return true
}

func (p *Parser) Err() error {
	return nil
}

func (p *Parser) Short(toMatch string) string {
	if toMatch == "" {
		panic("argument to Short must not be empty")
	}

	if p.state != short {
		return ""
	}

	return toMatch
}

func (p *Parser) Long(toMatch string) string {
	if toMatch == "" {
		panic("argument to Long must not be empty")
	}

	if !(p.state == empty || p.state == pendingValue) {
		return ""
	}

	return toMatch
}
