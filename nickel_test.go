package nickel

import (
	"bytes"
	"strings"
	"testing"
)

func TestRecord(t *testing.T) {
	ctx := NewContext()
	expr, err := ctx.EvalDeep("{ foo = 1, bar = 2 }")

	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	record, ok := expr.ToRecord()
	if !ok {
		t.Fatal("not a record")
	}

	elts := record.Elements()
	if len(elts) != 2 {
		t.Fatal("expected 2 elements")
	}
	_, ok = elts["foo"]
	if !ok {
		t.Fatal("no foo")
	}
	_, ok = elts["bar"]
	if !ok {
		t.Fatal("no bar")
	}
}

func TestEvalError(t *testing.T) {
	ctx := NewContext()
	_, err := ctx.EvalDeep("{ foo | String = 1, bar = 2 }")

	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "broken by the value of `foo`") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

type FooBar struct {
	Foo int `json:"foo"`
	Bar int `json:"bar"`
}

func TestMarshalJSON(t *testing.T) {
	ctx := NewContext()
	expr, err := ctx.EvalDeep("{ foo | Number = 1, bar = 2 }")

	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	_, err = expr.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var target FooBar
	err = expr.ConvertTo(&target)
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}
	if target.Foo != 1 {
		t.Fatalf("expected foo = 1")
	}
	if target.Bar != 2 {
		t.Fatalf("expected bar = 2")
	}
}

func TestTrace(t *testing.T) {
	var buf bytes.Buffer

	ctx := NewContext()
	ctx.SetTraceWriter(&buf)
	_, err := ctx.EvalDeep("std.trace \"hi\" { bye = std.trace \"bye\" 1 }")
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	traceOutput := buf.String()
	if traceOutput != "std.trace: hi\nstd.trace: bye\n" {
		t.Fatalf("unexpected buf contents: `%s`", traceOutput)
	}
}
