package execx_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

func TestOSOutputReturnsStdout(t *testing.T) {
	out, err := execx.OS{}.Output("", "sh", "-c", "echo hi")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "hi\n" {
		t.Fatalf("got %q", out)
	}
}

func TestOSOutputErrorIncludesStderr(t *testing.T) {
	_, err := execx.OS{}.Output("", "sh", "-c", "echo boom >&2; exit 3")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("want error containing stderr, got %v", err)
	}
}

func TestOSInteractiveWritesStdout(t *testing.T) {
	var buf bytes.Buffer
	if err := (execx.OS{}).Interactive(t.TempDir(), &buf, "sh", "-c", "echo hi"); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "hi\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestFakeRecordsCallsAndWritesInteractiveOutput(t *testing.T) {
	f := &execx.Fake{Handler: func(c execx.Call) ([]byte, error) {
		if c.Line() == "wtm resolve" {
			return []byte("/wt/a\n"), nil
		}
		return nil, nil
	}}
	var buf bytes.Buffer
	if err := f.Interactive("/repo", &buf, "wtm", "resolve"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Output("/repo", "wtm", "list", "--output", "json"); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "/wt/a\n" {
		t.Fatalf("stdout %q", buf.String())
	}
	want := []string{"wtm resolve", "wtm list --output json"}
	if strings.Join(f.Lines(), "|") != strings.Join(want, "|") {
		t.Fatalf("lines %v", f.Lines())
	}
	if !f.Calls[0].Interactive || f.Calls[1].Interactive || f.Calls[0].Dir != "/repo" {
		t.Fatalf("calls %+v", f.Calls)
	}
}
