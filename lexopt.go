package lexopt

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrNoToken         = fmt.Errorf("no token available")
	ErrUnexpectedValue = fmt.Errorf("unexpected value")
	ErrNoValue         = fmt.Errorf("no value found")
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
			p.state = empty
			break
		}

		if strings.HasPrefix(p.short, "=") && p.short != "=" {
			p.err = ErrUnexpectedValue
			return false
		}

		p.Current = toShort(p.short[0])
		p.short = p.short[1:]

		if p.short == "" {
			p.state = empty
		}
		return true

	case finished:
		nextTok, err := p.nextTok()
		switch {
		case errors.Is(err, ErrNoToken):
			return false
		case err != nil:
			p.err = err
			return false
		default:
			p.Current = toPositional(nextTok)
			return true
		}
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
			p.Current = toPositional(nextTok)
			return true
		}

		if len(nextTok) > 2 {
			p.state = short
			p.short = nextTok[2:]
		}

		p.Current = toShort(nextTok[1])
		return true

	default:
		p.Current = toPositional(nextTok)
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
			return noMatch(), ErrNoValue
		}

		return toValue(val), nil

	case short:
		if p.short == "=" {
			// If you passed -x= and want a value, that's silly.
			return noMatch(), ErrNoValue
		}

		// Remove a leading equals, if we have it, and then return everything
		// else.
		val := toValue(strings.TrimPrefix(p.short, "="))
		p.short = ""
		p.state = empty
		return val, nil

	case finished:
		return noMatch(), ErrNoValue

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

func (p *Parser) dumpState(w io.Writer) {
	fmt.Fprintln(w, "--- parser state ---")
	fmt.Fprintln(w, "Current:    ", p.Current)
	fmt.Fprintln(w, "argv:       ", p.argv)
	fmt.Fprintln(w, "idx:        ", p.idx)
	fmt.Fprintln(w, "state:      ", p.state)
	fmt.Fprintln(w, "pending arg:", p.pending)
	fmt.Fprintln(w, "short:      ", p.short)
	fmt.Fprintln(w, "err:        ", p.err)
	fmt.Fprintln(w, "---")
}
