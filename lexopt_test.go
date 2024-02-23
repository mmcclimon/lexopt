package lexopt

import (
	"testing"
)

type parserTester struct {
	*Parser
}

func newTester(argv ...string) *parserTester {
	return &parserTester{New(argv)}
}

func (pt *parserTester) nextOk(t *testing.T) {
	if !pt.Next() {
		t.Error(".Next() returned false, expected value")
	}
}

func (pt *parserTester) emptyOk(t *testing.T) {
	if pt.Next() {
		t.Error(".Next() returned true, expected empty")
	}
}

func TestSingleLongOpt(t *testing.T) {
	pt := newTester("--foo")
	pt.nextOk(t)

	if pt.Current != pt.Long("foo") {
		t.Error(`--foo should match p.Long("foo")`)
	}

	pt.emptyOk(t)
}

func TestNoOptions(t *testing.T) {
	pt := newTester()
	pt.emptyOk(t)
}

func TestDoubleDash(t *testing.T) {
	pt := newTester("--foo", "--", "whatever")
	pt.nextOk(t)
	pt.emptyOk(t)
	pt.emptyOk(t)
}

func TestPanics(t *testing.T) {
	pt := newTester()

	tests := map[string]func(string) Arg{
		"Short": pt.Short,
		"Long":  pt.Long,
	}

	for desc, method := range tests {
		t.Run(desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("%s did not panic on empty string", desc)
				}
			}()

			method("")
		})
	}
}
