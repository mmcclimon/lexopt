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
		// We have an --long=value with an unconsumed value; this is an error.
		p.err = ErrUnexpectedValue
		return false

	case short:
		// We have an -s=value with an unconsumed value; this is an error.
		if strings.HasPrefix(p.short, "=") && p.short != "=" {
			p.err = ErrUnexpectedValue
			return false
		}

		// Take the next short option out of an -abc set.
		p.Current = toShort(p.short[0])
		p.updateShort(p.short[1:])
		return true

	case finished:
		nextTok, err := p.nextTok()
		if err != nil {
			return false
		}

		p.Current = toPositional(nextTok)
		return true
	}

	if p.state != empty {
		panic("unexpected state")
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

		p.Current = toShort(nextTok[1])
		p.updateShort(nextTok[2:])
		return true

	default:
		p.Current = toPositional(nextTok)
		return true
	}
}

// this should return an error
func (p *Parser) Value() (Arg, error) {
	switch p.state {
	case pendingValue:
		val := toValue(p.pending)
		p.state = empty
		p.pending = ""
		return val, nil

	case empty:
		val, err := p.nextTok()
		if err != nil {
			return noMatch(), ErrNoValue
		}

		return toValue(val), nil

	case short:
		// Remove a leading equals, if we have it, and then return everything
		// else.
		val := toValue(strings.TrimPrefix(p.short, "="))
		p.updateShort("")
		return val, nil

	case finished:
		return noMatch(), ErrNoValue

	default:
		panic("unreachable")
	}
}

func (p *Parser) Err() error {
	return p.err
}

func (p *Parser) nextTok() (string, error) {
	if p.idx >= len(p.argv) {
		return "", ErrNoToken
	}

	next := p.argv[p.idx]
	p.idx++
	return next, nil
}

// Set p.short to remaining and update the state correctly for empty string.
func (p *Parser) updateShort(remaining string) {
	p.short = remaining

	if remaining == "" {
		p.state = empty
	} else {
		p.state = short
	}
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
