package lexopt

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoToken         = fmt.Errorf("no token available")
	ErrUnexpectedValue = fmt.Errorf("unexpected value")
)

type Parser struct {
	Current Arg

	argv    []string
	idx     int
	state   state
	pending string
	short   string
	err     error
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
	switch p.state {
	case pendingValue:
		p.err = ErrUnexpectedValue
		return false

	case short:
		if len(p.short) == 0 {
			// TODO: should we set this here or elsewhere?
			p.state = empty
			break
		}

		if len(p.short) > 1 && strings.HasPrefix(p.short, "=") {
			p.err = ErrUnexpectedValue
			return false
		}

		p.Current = toShort(p.short[0])
		p.short = p.short[1:]
		return true

	case finished:
		nextTok, err := p.nextTok()
		if err != nil {
			p.err = err
			return false
		}

		p.Current = toValue(nextTok)
		return true
	}

	if p.state != empty {
		panic("unexpected state!")
	}

	nextTok, err := p.nextTok()
	if errors.Is(err, ErrNoToken) {
		return false
	}

	switch {
	case nextTok == "--":
		p.state = finished
		return p.Next()

	case strings.HasPrefix(nextTok, "--"):
		p.state = empty

		nextTok = strings.TrimPrefix(nextTok, "--")
		before, after, hasEqual := strings.Cut(nextTok, "=")
		if hasEqual {
			p.pending = after
			p.state = pendingValue
		}
		p.Current = toLong(before)
		return true

	case strings.HasPrefix(nextTok, "-"):
		if nextTok == "-" {
			p.Current = toValue(nextTok)
			return true
		}

		p.state = short
		p.Current = toShort(nextTok[1])
		p.short = strings.TrimPrefix(nextTok[2:], "-")
		return true

	default:
		p.Current = toValue(nextTok)
		return true
	}

}

func (p *Parser) Err() error {
	return p.err
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

// this should return an error
func (p *Parser) Value() (Arg, error) {
	switch p.state {
	case pendingValue:
		p.state = empty
		return toValue(p.pending), nil

	case empty:
		val, err := p.nextTok()
		if err != nil {
			return noMatch(), err
		}

		return toValue(val), nil

	case short:
		if p.short == "=" {
			// -x= is nonsense, sorry.
			return noMatch(), ErrNoToken
		}

		// Here, we're asking for a value for a short option, so we can just
		// return everything we haven't yet consumed.

		p.short = strings.TrimPrefix(p.short, "=")

		val := toValue(p.short)
		p.short = ""
		p.state = empty
		return val, nil

	case finished:
		return p.Current, nil

	default:
		panic("unreachable")
	}

}

func (p *Parser) nextTok() (string, error) {
	if p.idx >= len(p.argv) {
		return "", ErrNoToken
	}

	next := p.argv[p.idx]
	p.idx++
	return next, nil
}
