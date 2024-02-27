package lexopt_test

import (
	"fmt"
	"strings"

	"github.com/mmcclimon/lexopt"
)

func ExampleParser() {
	var (
		number   int
		loud     bool
		greeting string
	)

	parser := lexopt.New([]string{"myapp", "-n=3", "--loud", "hello"})
	for parser.Next() {
		switch arg := parser.Current; arg {
		case lexopt.Short('n'):
			n, _ := parser.Value()
			number = n.MustInt()
		case lexopt.Long("loud"):
			loud = true
		default:
			greeting = arg.String()
		}
	}

	for i := 0; i < number; i++ {
		if loud {
			fmt.Println(strings.ToUpper(greeting))
		} else {
			fmt.Println(greeting)
		}
	}
	// Output:
	// HELLO
	// HELLO
	// HELLO
}

func ExampleParser_Next() {
	parser := lexopt.NewFromArgs([]string{"foo", "bar", "baz"})
	for parser.Next() {
		fmt.Println(parser.Current)
	}

	fmt.Println(parser.Err())
	// OUTPUT:
	// foo
	// bar
	// baz
	// <nil>
}

func ExampleParser_Value() {
	parser := lexopt.NewFromArgs([]string{"-f=pathname"})
	for parser.Next() {
		if parser.Current == lexopt.Short('f') {
			val, _ := parser.Value()
			fmt.Println(val)
		}
	}
	// OUTPUT: pathname
}

func ExampleParser_OptionalValue() {
	parser := lexopt.NewFromArgs([]string{"-f=pathname", "-o", "file"})
	parser.Next()
	val, ok := parser.OptionalValue()
	fmt.Printf("%s, %v, %s\n", parser.Current, ok, val)

	parser.Next()
	_, ok = parser.OptionalValue()
	fmt.Printf("%s, %v\n", parser.Current, ok)

	// OUTPUT:
	// f, true, pathname
	// o, false
}

func ExampleParser_Values() {
	parser := lexopt.NewFromArgs([]string{"--command", "echo", "Hello world", "-q"})
	parser.Next()
	vals, _ := parser.Values()
	fmt.Printf("%s: %q\n", parser.Current, vals)
	// OUTPUT:
	// command: ["echo" "Hello world"]
}

func ExampleRawArgs() {
	parser := lexopt.NewFromArgs([]string{"-c", "file", "-o", "output", "stop", "-q"})
	args, _ := parser.RawArgs()
	for args.Next() {
		fmt.Println("arg:", args.Current)
		if args.Current.String() == "stop" {
			break
		}
	}

	parser.Next()
	fmt.Println("parser:", parser.Current)
	// OUTPUT:
	// arg: -c
	// arg: file
	// arg: -o
	// arg: output
	// arg: stop
	// parser: q
}

func ExampleRawArgs_NextIf() {
	parser := lexopt.NewFromArgs([]string{"-o", "output", "stop", "-q"})
	predicate := func(a lexopt.Arg) bool { return !strings.Contains(a.String(), "stop") }
	args, _ := parser.RawArgs()
	fmt.Println(args.NextIf(predicate))
	fmt.Println(args.NextIf(predicate))
	fmt.Println(args.NextIf(predicate))

	parser.Next()
	fmt.Println("parser:", parser.Current)
	// OUTPUT:
	// -o true
	// output true
	//  false
	// parser: stop
}

func ExampleRawArgs_AsStringSlice() {
	parser := lexopt.NewFromArgs([]string{"-o", "output", "-q"})
	args, _ := parser.RawArgs()
	fmt.Printf("%v\n", args.AsStringSlice())
	fmt.Println(args.Next())
	// OUTPUT:
	// [-o output -q]
	// false
}
