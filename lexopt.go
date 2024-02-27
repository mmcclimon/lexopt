// Package lexopt provides a very simple command line argument parser. Rather
// than declaring flags up front, lexopt provides a parser and a stream of
// options, allowing you to decide what to do with each token.
package lexopt

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	// ErrUnexpectedValue is returned when an argument has a value that was not
	// consumed by a call to parser.Value().
	ErrUnexpectedValue = fmt.Errorf("unexpected value")

	// ErrNoValue is returned when a value is expected, but not available.
	ErrNoValue = fmt.Errorf("no value found")

	// errNoToken is internal-only, raised when parser.argv is exhausted.
	errNoToken = fmt.Errorf("no token available")
)

// Parser is a parser for command line arguments.
type Parser struct {
	// Current is the parser's current argument. The Next method advances the
	// Current argument.
	Current Arg

	// internal state

	binName  string   // $0, possibly empty
	argv     []string // original args, not including binName
	idx      int      // index into argv
	state    state    // current parser state
	pending  string   // when state=pendingValue, the value that's pending
	short    []rune   // when state=short, the code points for the current arg
	shortpos int      // index into short
	err      error    // can be set when Next() returns false
}

// The state type is used for storing the internal state of the parser.
type state int

const (
	empty        state = iota // no state, or maybe we just parsed a long opt
	short                     // we've started parsing a short option, and maybe there are more
	pendingValue              // we parsed --long=value, and waiting to yield value
	finished                  // we saw --
)

// New returns a new parser; fullArgv must contain the binary name as the
// first element.
func New(fullArgv []string) *Parser {
	return &Parser{
		binName: fullArgv[0],
		argv:    fullArgv[1:],
	}
}

// NewFromEnv returns a new parser instantiated with [os.Args]. It is a
// shorthand for New(os.Args).
func NewFromEnv() *Parser {
	return New(os.Args)
}

// NewFromArgs returns a new parser from the given args; argv must _not_
// contain the binary name.
func NewFromArgs(argv []string) *Parser {
	return &Parser{
		argv: argv,
	}
}

// BinName returns the program name from the command line, as in the first
// element of [os.Args]. If the parser was created with NewFromArgs, BinName
// always returns the empty string.
func (p *Parser) BinName() string {
	return p.binName
}

// Next advances Parser's internal iterator, possibly setting
// Parser.Current. It returns true if it successfully sets Parser.Current,
// false otherwise. After consuming all arguments with Next, you should check
// the result of [Parser.Err]. If it is non-nil, it contains the parse error
// that caused iteration to fail.
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

		p.Current = Value(nextTok)
		return true
	}

	if p.state != empty {
		panic("unexpected state")
	}

	nextTok, err := p.nextTok()
	if errors.Is(err, errNoToken) {
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
			p.Current = Value(nextTok)
			return true
		}

		p.resetShort(nextTok[1:])
		p.Current = p.takeShort()
		return true

	default:
		p.Current = Value(nextTok)
		return true
	}
}

// Value returns a value for an argument. This function should normally be
// called right after seeing an option that expects a value; positional
// arguments should be collected using [Parser.Next]. Note that this method will
// collect a value even if it looks like an option (i.e., it starts with -).
func (p *Parser) Value() (Arg, error) {
	ret, _, err := p.value()
	return ret, err
}

// OptionalValue returns a value only if it’s concatenated to an option, as in
// -ovalue or --option=value or -o=value, but not -o value or --option value.
func (p *Parser) OptionalValue() (Arg, bool) {
	if !p.hasPending() {
		return Arg{}, false
	}

	ret, _, _ := p.value()
	return ret, true
}

// value is the internal implementation for Value and OptionalValue. The
// middle boolean return is whether or not there was an equals sign (which
// matters for Values).
func (p *Parser) value() (Arg, bool, error) {
	switch p.state {
	case pendingValue:
		val := Value(p.pending)
		p.state = empty
		p.pending = ""
		return val, true, nil

	case empty:
		val, err := p.nextTok()
		if err != nil {
			return Arg{}, false, ErrNoValue
		}

		return Value(val), false, nil

	case short:
		// Remove a leading equals, if we have it, and then return everything else.
		raw := string(p.short[p.shortpos:])
		hasEqual := strings.HasPrefix(raw, "=")
		val := Value(strings.TrimPrefix(raw, "="))
		p.resetShort("")
		return val, hasEqual, nil

	case finished:
		return Arg{}, false, ErrNoValue

	default:
		panic("unreachable")
	}
}

// Values gathers multiple values for an option.  This is used for options
// that take multiple arguments, such as a --command flag that’s invoked as
// app --command echo 'Hello world'.
//
// It will gather arguments until another option is found, or -- is found, or
// the end of the command line is reached. This differs from .value(), which
// takes a value even if it looks like an option.
//
// An equals sign (=) will limit this to a single value. That means -a=b c and
// --opt=b c will only yield "b" while -a b c, -ab c and --opt b c will yield
// "b", "c".
//
// If not at least one value is found then it returns [ErrNoValue].
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
		vals = append(vals, Value(val))
	}

	return vals, nil
}

// Err returns the last error seen by [Parser.Next]. If argument processing
// ended normally, Err returns nil.
func (p *Parser) Err() error {
	return p.err
}

// nextTok advances the internal iterator, returning the next token.
func (p *Parser) nextTok() (string, error) {
	if p.idx >= len(p.argv) {
		return "", errNoToken
	}

	next := p.argv[p.idx]
	p.idx++
	return next, nil
}

// takeShort returns the next short option out of p.short, updating the
// internal state as required.
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

// resetShort sets p.short to a rune slice of value, updating the internal
// state as required.
func (p *Parser) resetShort(value string) {
	p.short = []rune(value)
	p.shortpos = 0

	if value == "" {
		p.state = empty
	} else {
		p.state = short
	}
}

// hasPending returns true if we know there is a value pending (i.e., if
// OptionalValue would return true).
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

// nextIsNormal returns true if the next token is a non-option.
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

// dumpState writes the parser state to out, which defaults to os.Stdout. It's
// useful for debugging.
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

// RawArgs takes raw arguments from the middle of the original command line.
// The return value is a [RawArgs] struct, which can be used as an iterator.
//
// RawArgs returns [ErrUnexpectedValue] if the last option had a left-over
// argument, as in --option=value, -ovalue, or if it was midway through an
// option chain, as in -abc.
func (p *Parser) RawArgs() (*RawArgs, error) {
	if p.hasPending() {
		return nil, ErrUnexpectedValue
	}
	return &RawArgs{parser: p}, nil
}

// RawArgs is an iterator over the raw arguments for a parser; it is returned
// by the [Parser.RawArgs] method. It shares state with the parser, so it's
// possible to consume some of the arguments, then return control to the main
// parser.
type RawArgs struct {
	// Current is the current argument. The Next method advances the Current argument.
	Current Arg
	parser  *Parser
}

// Next advances Parser's internal iterator, possibly setting
// Parser.Current. It returns true if it successfully sets Parser.Current,
// and false when the command line is exhausted. Unlike [Parser.Next], Next
// will never result in an error (and consequently, RawArgs does not have an
// Err method).
func (ra *RawArgs) Next() bool {
	nextTok, err := ra.parser.nextTok()
	if err != nil {
		return false
	}

	ra.Current = Value(nextTok)
	return true
}

// NextIf returns the next raw argument, only if predicate is true.
func (ra *RawArgs) NextIf(predicate func(Arg) bool) (Arg, bool) {
	// This is pretty unidiomatic in Go, but I'm implementing it for the compat
	// tests.
	arg, ok := ra.Peek()
	if !ok {
		return Arg{}, false
	}

	if predicate(arg) {
		ra.Next()
		return arg, true
	}

	return Arg{}, false
}

// Peek returns the next raw argument but does not set RawArgs.Current. It
// returns false if the arguments have been exhausted.
func (ra *RawArgs) Peek() (Arg, bool) {
	if ra.parser.idx >= len(ra.parser.argv) {
		return Arg{}, false
	}

	return Value(ra.parser.argv[ra.parser.idx]), true
}

// AsSlice returns all of the raw arguments as an Args slice. This method
// exhausts the iterator; after a call to AsSlice, all further calls to
// [RawArgs.Next] will return false.
func (ra *RawArgs) AsSlice() []Arg {
	args := make([]Arg, 0, len(ra.parser.argv)-ra.parser.idx)
	for ra.Next() {
		args = append(args, ra.Current)
	}
	return args
}

// AsSlice returns all of the raw arguments as a slice of strings. This method
// exhausts the iterator; after a call to AsStringSlice, all further calls to
// [RawArgs.Next] will return false.
func (ra *RawArgs) AsStringSlice() []string {
	args := make([]string, 0, len(ra.parser.argv)-ra.parser.idx)
	for ra.Next() {
		args = append(args, ra.Current.String())
	}
	return args
}
