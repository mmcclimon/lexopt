package lexopt

// All the tests in this file are taken directly from the Rust lexopt tests.

import "testing"

func TestBasicCompat(t *testing.T) {
	pt := newTester(t, "-n", "10", "foo", "-", "--", "baz", "-qux")
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
	pt := newTester(t, "-abc", "-fvalue", "-xfvalue")
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
	pt := newTester(t, "--foo", "--bar=qux", "--foobar=qux=baz")
	pt.longOk("foo")
	pt.longOk("bar")
	pt.valueOk("qux")
	pt.longOk("foobar")
	pt.nextErrOk(ErrUnexpectedValue)
	pt.emptyOk()
}

func TestDashArgsCompat(t *testing.T) {
	t.Run("double dash end of options", func(t *testing.T) {
		pt := newTester(t, "-x", "--", "-y")
		pt.shortOk('x')
		pt.positionalOk("-y")
		pt.emptyOk()
	})

	t.Run("but not as value", func(t *testing.T) {
		pt := newTester(t, "-x", "--", "-y")
		pt.shortOk('x')
		pt.valueOk("--")
		pt.shortOk('y')
		pt.emptyOk()
	})

	t.Run("dash is valid value", func(t *testing.T) {
		pt := newTester(t, "-x", "-", "-y")
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

	pt3 := newTester(t)
	pt3.noValueOk()
}

func TestWeirdValuesCompat(t *testing.T) {
	pt := newTester(t, "", "--=", "--=3", "-", "-x", "--", "-", "-x", "--", "", "-", "-x")
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
	pt := newTester(t, "-aµ", "--µ=10", "µ", "--foo=µ")
	pt.shortOk('a')
	pt.shortOk('µ')
	pt.longOk("µ")
	pt.valueOk("10")
	pt.positionalOk("µ")
	pt.longOk("foo")
	pt.valueOk("µ")
}
