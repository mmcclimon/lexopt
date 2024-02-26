package lexopt

// All the tests in this file are taken directly from the Rust lexopt tests.

import "testing"

func TestBasicCompat(t *testing.T) {
	pt := newTester(t, "-n 10 foo - -- baz -qux")
	pt.shortOk('n')
	pt.valueOk("10")
	pt.positionalOk("foo")
	pt.positionalOk("-")
	pt.positionalOk("baz")
	pt.positionalOk("-qux")
	pt.emptyOk()
	pt.emptyOk()
	pt.emptyOk()
}

func TestCombinedCompat(t *testing.T) {
	pt := newTester(t, "-abc -fvalue -xfvalue")
	pt.shortOk('a')
	pt.shortOk('b')
	pt.shortOk('c')
	pt.shortOk('f')
	pt.valueOk("value")
	pt.shortOk('x')
	pt.shortOk('f')
	pt.valueOk("value")
	pt.emptyOk()
}

func TestLongCompat(t *testing.T) {
	pt := newTester(t, "--foo --bar=qux --foobar=qux=baz")
	pt.longOk("foo")
	pt.longOk("bar")
	pt.valueOk("qux")
	pt.longOk("foobar")
	pt.nextErrOk(ErrUnexpectedValue)
	pt.emptyOk()
}

func TestDashArgsCompat(t *testing.T) {
	t.Run("double dash end of options", func(t *testing.T) {
		pt := newTester(t, "-x -- -y")
		pt.shortOk('x')
		pt.positionalOk("-y")
		pt.emptyOk()
	})

	t.Run("but not as value", func(t *testing.T) {
		pt := newTester(t, "-x -- -y")
		pt.shortOk('x')
		pt.valueOk("--")
		pt.shortOk('y')
		pt.emptyOk()
	})

	t.Run("dash is valid value", func(t *testing.T) {
		pt := newTester(t, "-x - -y")
		pt.shortOk('x')
		pt.positionalOk("-")
		pt.shortOk('y')
		pt.emptyOk()
	})

	// As the Rust tests say: '-' is a silly and hard to use short option, but
	// other parsers treat it like an option in this position.
	t.Run("dash as short option", func(t *testing.T) {
		pt := newTester(t, "-x-y")
		pt.shortOk('x')
		pt.shortOk('-')
		pt.shortOk('y')
		pt.emptyOk()
	})
}

func TestMissingValuesCompat(t *testing.T) {
	pt := newTester(t, "-o")
	pt.shortOk('o')
	pt.noValueOk()

	pt2 := newTester(t, "--out")
	pt2.longOk("out")
	pt2.noValueOk()

	pt3 := newTester(t, "")
	pt3.noValueOk()
}

func TestWeirdValuesCompat(t *testing.T) {
	pt := newTesterArgs(t, "", "--=", "--=3", "-", "-x", "--", "-", "-x", "--", "", "-", "-x")
	pt.positionalOk("")

	// Weird and questionable, indeed!
	pt.longOk("")
	pt.valueOk("")
	pt.longOk("")
	pt.valueOk("3")

	pt.positionalOk("-")
	pt.shortOk('x')
	pt.valueOk("--")

	pt.positionalOk("-")
	pt.shortOk('x')

	pt.positionalOk("")
	pt.positionalOk("-")
	pt.positionalOk("-x")

	pt.emptyOk()
}

func TestShortOptEqualsSignCompat(t *testing.T) {
	pt := newTester(t, "-a=b")
	pt.shortOk('a')
	pt.valueOk("b")

	pt = newTester(t, "-a=b")
	pt.shortOk('a')
	pt.nextErrOk(ErrUnexpectedValue)
	pt.emptyOk()

	pt = newTester(t, "-a=")
	pt.shortOk('a')
	pt.valueOk("")
	pt.emptyOk()

	pt = newTester(t, "-a=")
	pt.shortOk('a')
	pt.nextErrOk(ErrUnexpectedValue)
	pt.emptyOk()

	pt = newTester(t, "-=")
	pt.shortOk('=')
	pt.emptyOk()

	pt = newTester(t, "-=a")
	pt.shortOk('=')
	pt.valueOk("a")
}

func TestUnicodeCompat(t *testing.T) {
	pt := newTester(t, "-aµ --µ=10 µ --foo=µ")
	pt.shortOk('a')
	pt.shortOk('µ')
	pt.longOk("µ")
	pt.valueOk("10")
	pt.positionalOk("µ")
	pt.longOk("foo")
	pt.valueOk("µ")
}

func TestMultiValuesCompat(t *testing.T) {
	for _, test := range []string{"-a b c d", "-ab c d", "-a b c d --", "--a b c d"} {
		pt := newTester(t, test)
		pt.nextOk()
		pt.valuesOk("b", "c", "d")
	}

	for _, test := range []string{"-a=b c", "--a=b c"} {
		pt := newTester(t, test)
		pt.nextOk()
		pt.valuesOk("b")
		pt.positionalOk("c")
	}

	for _, test := range []string{"-a", "--a", "-a -b", "-a -- b", "-a --"} {
		pt := newTester(t, test)
		pt.nextOk()
		pt.noValuesOk()
		pt.Next()

		if err := pt.Err(); err != nil {
			t.Errorf(".Err() returned unexpected err: %s", err)
		}
	}

	for _, test := range []string{"-a=", "--a="} {
		pt := newTester(t, test)
		pt.nextOk()
		pt.valuesOk("")
		pt.emptyOk()
	}

	// Rust lexopt has tests here that assert that leaving the values iterator
	// does not change the internal state of the parser. We can't do that in Go,
	// because we don't have iterators. The tests are included here to
	// demonstrate the differences.

	t.Run("rust lexopt incompatibility", func(t *testing.T) {
		t.Skip("go has no iteration protocol")

		for _, test := range []string{"-a=b", "--a=b", "-a b"} {
			pt := newTester(t, test)
			pt.nextOk()
			pt.valuesOk()
			pt.valueOk("b")
		}

		pt := newTester(t, "-ab")
		pt.shortOk('a')
		pt.valuesOk()
		pt.shortOk('b')
	})
}

func TestRawArgsCompat(t *testing.T) {
	pt := newTester(t, "-a b c d")
	args := pt.rawArgsOk()
	args.argSliceOk("-a", "b", "c", "d")
	pt.rawArgsOk()
	pt.emptyOk()
	pt.rawArgsOk().emptyOk()

	pt = newTester(t, "-ab c d")
	pt.shortOk('a')
	pt.rawArgsErrOk()
	pt.valueOk("b")
	pt.rawArgsOk().stringSliceOk("c", "d")
	pt.emptyOk()
	pt.rawArgsOk().emptyOk()

	pt = newTester(t, "-a b c d")
	args = pt.rawArgsOk()
	args.nextArgOk("-a")
	args.nextArgOk("b")
	args.nextArgOk("c")
	pt.positionalOk("d")
	pt.emptyOk()

	pt = newTester(t, "a")
	args = pt.rawArgsOk()
	args.peekOk("a")

	isA := func(a Arg) bool { return a.String() == "a" }

	_, ok := args.NextIf(func(_ Arg) bool { return false })
	if ok {
		t.Fatalf("NextIf unexpectedly returned true")
	}

	arg, ok := args.NextIf(isA)
	if !ok {
		t.Fatalf("NextIf unexpectedly returned false")
	}
	if arg != toPositional("a") {
		t.Errorf("NextIf returned bad arg: %v", arg)
	}

	pt.emptyOk()

	_, ok = args.NextIf(isA)
	if ok {
		t.Fatalf("NextIf unexpectedly returned true")
	}

}
