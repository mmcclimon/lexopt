package lexopt

import (
	"errors"
	"fmt"
	"strings"
)

var ErrNoToken = fmt.Errorf("no token available")

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
	nextTok, err := p.nextTok()
	if errors.Is(err, ErrNoToken) {
		return false
	}

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

// this should return an error
func (p *Parser) Value() (string, error) {
	switch p.state {
	case pendingValue:
		p.state = empty
		return p.pending, nil

	case empty:
		return p.nextTok()
	}

	// TODO
	return "", nil
}

func (p *Parser) nextTok() (string, error) {
	if p.idx >= len(p.argv) || p.state == finished {
		return "", ErrNoToken
	}

	next := p.argv[p.idx]
	p.idx++
	return next, nil
}
