package lexopt

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrNoToken         = fmt.Errorf("no token available")
	ErrUnexpectedValue = fmt.Errorf("unexpected value")
	ErrNoValue         = fmt.Errorf("no value found")
)

type Parser struct {
	Current Arg

	argv     []string
	idx      int
	state    state
	pending  string
	short    []rune
	shortpos int
	err      error
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
		if p.short[p.shortpos] == '=' && p.shortpos >= 1 {
			p.err = ErrUnexpectedValue
			return false
		}

		// Take the next short option out of an -abc set.
		p.Current = p.takeShort()
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
		p.Current = Long(before)
		return true

	case strings.HasPrefix(nextTok, "-"):
		if nextTok == "-" {
			p.Current = toPositional(nextTok)
			return true
		}

		p.resetShort(nextTok[1:])
		p.Current = p.takeShort()
		return true

	default:
		p.Current = toPositional(nextTok)
		return true
	}
}

func (p *Parser) Value() (Arg, error) {
	ret, _, err := p.value()
	return ret, err
}

func (p *Parser) OptionalValue() (Arg, bool) {
	if !p.hasPending() {
		return Arg{}, false
	}

	ret, _, _ := p.value()
	return ret, true
}

func (p *Parser) value() (Arg, bool, error) {
	switch p.state {
	case pendingValue:
		val := toValue(p.pending)
		p.state = empty
		p.pending = ""
		return val, true, nil

	case empty:
		val, err := p.nextTok()
		if err != nil {
			return noMatch(), false, ErrNoValue
		}

		return toValue(val), false, nil

	case short:
		// Remove a leading equals, if we have it, and then return everything else.
		raw := string(p.short[p.shortpos:])
		hasEqual := strings.HasPrefix(raw, "=")
		val := toValue(strings.TrimPrefix(raw, "="))
		p.resetShort("")
		return val, hasEqual, nil

	case finished:
		return noMatch(), false, ErrNoValue

	default:
		panic("unreachable")
	}
}

func (p *Parser) Values() ([]Arg, error) {
	if !p.hasPending() && !p.nextIsNormal() {
		return nil, ErrNoValue
	}

	var vals []Arg

	// Take one.
	val, hadEqual, _ := p.value()
	vals = append(vals, val)

	// Take more, if we can.
	for !hadEqual && p.nextIsNormal() {
		val, _ := p.nextTok()
		vals = append(vals, toValue(val))
	}

	return vals, nil
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
func (p *Parser) takeShort() Arg {
	ret := Short(p.short[p.shortpos])
	p.shortpos++

	if p.shortpos >= len(p.short) {
		p.state = empty
		p.short = nil
		p.shortpos = 0
	} else {
		p.state = short
	}

	return ret
}

func (p *Parser) resetShort(value string) {
	p.short = []rune(value)
	p.shortpos = 0

	if value == "" {
		p.state = empty
	} else {
		p.state = short
	}
}

func (p *Parser) hasPending() bool {
	switch p.state {
	case empty, finished:
		return false
	case pendingValue:
		return true
	case short:
		return p.shortpos < len(p.short)
	default:
		panic("unreachable")
	}
}

func (p *Parser) nextIsNormal() bool {
	if p.idx >= len(p.argv) {
		// out of options
		return false
	}

	next := p.argv[p.idx]

	switch {
	case p.state == finished:
		return true

	case next == "-":
		return true

	default:
		return !strings.HasPrefix(next, "-")
	}
}

func (p *Parser) dumpState(out ...io.Writer) {
	var w io.Writer = os.Stdout
	if len(out) > 0 {
		w = out[0]
	}

	fmt.Fprintln(w, "--- parser state ---")
	fmt.Fprintln(w, "Current:    ", p.Current)
	fmt.Fprintln(w, "argv:       ", p.argv)
	fmt.Fprintln(w, "idx:        ", p.idx)
	fmt.Fprintln(w, "state:      ", p.state)
	fmt.Fprintln(w, "pending arg:", p.pending)
	fmt.Fprintln(w, "short:      ", p.short)
	fmt.Fprintln(w, "shortpos:   ", p.shortpos)
	fmt.Fprintln(w, "err:        ", p.err)
	fmt.Fprintln(w, "---")
}
