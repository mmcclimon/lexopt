# lexopt

This is a port of [Rust's lexopt](https://github.com/blyxxyz/lexopt) to Go.
It aims to be a very faithful port, modulo the differences between Rust and
Go.

---

Lexopt is an argument parser for Go. It tries to have the simplest possible
design that's still correct. Much like Go itself, it's so simple that it's a
bit tedious to use.

Lexopt is:
- Small: one package, no dependencies. Easy to audit or vendor.
- Correct: standard conventions are supported and ambiguity is avoided. Well
  tested.
- Imperative: options are returned as they are found, nothing is declared
  ahead of time.
- Minimalist: only basic functionality is provided.
- Unhelpful: there is no help generation and error messages often lack context.

## Example

```go
type Args struct {
	thing  string
	number int
	shout  bool
}

func parseArgs() (Args, error) {
	args := Args{number: 1}
	parser := lexopt.NewFromEnv()

	for parser.Next() {
		switch arg := parser.Current; arg {
		case lexopt.Short('n'), lexopt.Long("number"):
			value, err := parser.Value()
			if err != nil {
				return args, err
			}
			args.number = value.MustInt()
		case lexopt.Long("shout"):
			args.shout = true
		case lexopt.Long("help"):
			fmt.Println("Usage: hello [-n|--number=NUM] [--shout] THING")
			os.Exit(0)
		default:
			if args.thing == "" {
				args.thing = arg.String()
			} else {
				return args, fmt.Errorf("unexpected arg: %s", arg.DashedString())
			}
		}
	}

	return args, nil
}

func main() {
	args, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	message := fmt.Sprintf("Hello %s", args.thing)
	if args.shout {
		message = strings.ToUpper(message)
	}

	for i := 0; i < args.number; i++ {
		fmt.Println(message)
	}
}
```

Let's walk through this:
- We start parsing with `lexopt.NewFromEnv()`.
- We call `parser.Next()` in a loop to get all the arguments until they run out.
- We match on arguments. `Short` and `Long` indicate an option.
- To get the value that belongs to an option (like `10` in `-n 10`) we call `parser.Value()`.
  - This returns an `Arg` type. `Arg` has methods for converting to common Go
    types.
  - Calling `parser.Value()` is how we tell `Parser` that `-n` takes a value at all.
- `Value` indicates a free-standing argument.
    - The `.String()` method decodes it into a plain `string`.
- If we don't know what to do with an argument, we return an error.

This covers most of the functionality in the library. Lexopt does very little for you.

## Command line syntax

The following conventions are supported:
- Short options (`-q`)
- Long options (`--verbose`)
- `--` to mark the end of options
- `=` to separate options from values (`--option=value`, `-o=value`)
- Spaces to separate options from values (`--option value`, `-o value`)
- Unseparated short options (`-ovalue`)
- Combined short options (`-abc` to mean `-a -b -c`)
- Options with optional arguments (like GNU sed's `-i`, which can be used standalone or as `-iSUFFIX`) (`Parser.OptionalValue()`)
- Options with multiple arguments (`Parser.Values()`)

These are not supported out of the box:
- Single-dash long options (like find's `-name`), or Go's standard flags.
- Abbreviated long options (GNU's getopt lets you write `--num` instead of `--number` if it can be expanded unambiguously)

`Parser.RawArgs()` provides an escape hatch for consuming the original command
line. This can be used for custom syntax, like treating `-123` as a number
instead of a string of options.

## Unicode

This library is not as pedantic as its Rust parent, because Go is inherently
less pendantic than Rust about these things. (In Go, a string is an arbitrary
collection of bytes; in Rust, strings are by definition valid UTF-8.) This
library deals almost exclusively in Go-standard `string`s, leaving the
encoding/decoding to the user, if required.

Short options may be unicode, but only a single codepoint (a `rune`).

## Why?

I noticed the Rust lexopt when [ripgrep](https://github.com/BurntSushi/ripgrep)
switched to it, and thought it would be interesting to port to Go.

This library may also be useful if a lot of control is desired, like when the
exact argument order matters or not all options are known ahead of time. It
could be considered more of a lexer than a parser.

## Why not?

This library may not be worth using if:
- You don't care about exact compliance and correctness
- You don't care about code size
- You do care about great error messages
- You hate boilerplate

## Differences from the Rust version

Rust and Go are different, and Rust is generally much nicer than Go. To that
end, there are a few notable differences in this port:

- The match syntax is not available (alas), so you generally need to match
  positional arguments with a switch/default, rather than enum destructuring.
- Go does not have iterators (alas), though Parser and RawArgs are half-baked
  versions of them.
- As an upshot of the above: the Rust `.values()` method returns an iterator
  that does not consume parser arguments unless you use the iterator. In the
  Go version, we simply return a slice instead.

## See also

- [The original lexopt](https://github.com/blyxxyz/lexopt)

