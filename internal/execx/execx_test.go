package execx_test

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

func sh(script string) execx.Cmd {
	return execx.Cmd{Name: "sh", Args: []string{"-c", script}}
}

func TestOSOutputPassesEnv(t *testing.T) {
	c := sh(`printf %s "$HERDR_WTM_TEST"`)
	c.Env = []string{"HERDR_WTM_TEST=yes"}
	out, err := execx.OS{}.Output(c)
	if err != nil || string(out) != "yes" {
		t.Fatalf("out %q err %v", out, err)
	}
}

func TestOSOutputErrorCarriesStderrAndCode(t *testing.T) {
	_, err := execx.OS{}.Output(sh("echo boom >&2; exit 12"))
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err %v", err)
	}
	if code, ok := execx.ExitCode(err); !ok || code != 12 {
		t.Fatalf("code %d ok %v", code, ok)
	}
}

func TestOSStreamDeliversLines(t *testing.T) {
	var got []string
	err := execx.OS{}.Stream(context.Background(), sh(`printf 'a\nb\n'`), func(l []byte) { got = append(got, string(l)) })
	if err != nil || !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestOSStreamExitCode(t *testing.T) {
	err := execx.OS{}.Stream(context.Background(), sh("exit 20"), func([]byte) {})
	if code, ok := execx.ExitCode(err); !ok || code != 20 {
		t.Fatalf("code %d ok %v err %v", code, ok, err)
	}
}

func TestOSStreamStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	err := execx.OS{}.Stream(ctx, sh("echo ready; sleep 10"), func([]byte) { cancel() })
	if err == nil {
		t.Fatal("want the killed process's error")
	}
}

func TestOSInteractiveWritesToStdout(t *testing.T) {
	var out bytes.Buffer
	c := sh("echo hi")
	c.Stdout = &out
	if err := (execx.OS{}).Interactive(c); err != nil || out.String() != "hi\n" {
		t.Fatalf("out %q err %v", out.String(), err)
	}
}

func TestExitCodeOfPlainError(t *testing.T) {
	if _, ok := execx.ExitCode(errors.New("x")); ok {
		t.Fatal("a plain error has no exit code")
	}
	if code, ok := execx.ExitCode(execx.ExitError{Code: 2}); !ok || code != 2 {
		t.Fatalf("code %d ok %v", code, ok)
	}
}

func TestFakeStreamSplitsLinesAndRecords(t *testing.T) {
	f := &execx.Fake{Handler: func(execx.Call) ([]byte, error) { return []byte("one\n\ntwo\n"), execx.ExitError{Code: 1} }}
	var got []string
	err := f.Stream(context.Background(), execx.Cmd{Name: "wtm", Args: []string{"events"}}, func(l []byte) { got = append(got, string(l)) })
	if !slices.Equal(got, []string{"one", "two"}) || err == nil {
		t.Fatalf("got %v err %v", got, err)
	}
	if !f.Calls[0].Streamed || f.Lines()[0] != "wtm events" {
		t.Fatalf("calls %+v", f.Calls)
	}
}
