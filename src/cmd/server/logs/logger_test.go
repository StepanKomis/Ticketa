package logs_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
)

func testConfig(t *testing.T, level string) *config.Config {
	t.Helper()
	cfg := config.Defaults()
	cfg.Logging.Dir = t.TempDir()
	cfg.Logging.Level = level
	return cfg
}

// newTestLogger creates a Logger backed by a temp directory and attaches a
// buffer to capture output for assertions.
func newTestLogger(t *testing.T, prefix string) (*logs.Logger, *bytes.Buffer) {
	t.Helper()
	l, err := logs.NewLogger(prefix, testConfig(t, "info"))
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	t.Cleanup(func() { l.Close() })

	buf := &bytes.Buffer{}
	l.AddWriter(buf)
	return l, buf
}

func TestNewLogger_ErrorWrapped(t *testing.T) {
	// A prefix with a path separator causes os.OpenFile to fail; the error
	// must be wrapped so callers can inspect it.
	_, err := logs.NewLogger("bad/prefix", testConfig(t, "info"))
	if err == nil {
		t.Fatal("expected error for invalid prefix")
	}
	if !strings.Contains(err.Error(), "failed to open log file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewLogger_PrefixIsUppercased(t *testing.T) {
	l, buf := newTestLogger(t, "myprefix")
	l.Info("check prefix")
	if !strings.Contains(buf.String(), "[MYPREFIX]") {
		t.Errorf("expected [MYPREFIX] in output, got: %s", buf.String())
	}
}

func TestInfo(t *testing.T) {
	l, buf := newTestLogger(t, "test")
	l.Info("hello world")
	if !strings.Contains(buf.String(), "[INFO] hello world") {
		t.Errorf("expected [INFO] hello world in output, got: %s", buf.String())
	}
}

func TestError(t *testing.T) {
	l, buf := newTestLogger(t, "test")
	l.Error("something went wrong")
	if !strings.Contains(buf.String(), "[ERROR] something went wrong") {
		t.Errorf("expected [ERROR] something went wrong in output, got: %s", buf.String())
	}
}

// TestWarn_Panics documents that Warn uses log.Panicln and therefore panics.
func TestWarn_Panics(t *testing.T) {
	l, _ := newTestLogger(t, "test")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Warn to panic (it uses log.Panicln)")
		}
	}()
	l.Warn("this panics")
}

func TestDebug_SuppressedByDefault(t *testing.T) {
	cfg := testConfig(t, "info")
	l, err := logs.NewLogger("test", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	buf := &bytes.Buffer{}
	l.AddWriter(buf)

	l.Debug("hidden message")
	if strings.Contains(buf.String(), "hidden message") {
		t.Error("Debug should be suppressed when level is not debug")
	}
}

func TestDebug_ShownWhenEnabled(t *testing.T) {
	cfg := testConfig(t, "debug")
	l, err := logs.NewLogger("test", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	buf := &bytes.Buffer{}
	l.AddWriter(buf)

	l.Debug("visible message")
	if !strings.Contains(buf.String(), "[DEBUG] visible message") {
		t.Errorf("expected [DEBUG] visible message in output, got: %s", buf.String())
	}
}

func TestDebug_LevelCheckIsCaseInsensitive(t *testing.T) {
	cfg := testConfig(t, "DEBUG")
	l, err := logs.NewLogger("test", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	buf := &bytes.Buffer{}
	l.AddWriter(buf)

	l.Debug("case insensitive")
	if !strings.Contains(buf.String(), "[DEBUG] case insensitive") {
		t.Errorf("expected [DEBUG] case insensitive in output, got: %s", buf.String())
	}
}

func TestInfof(t *testing.T) {
	l, buf := newTestLogger(t, "test")
	l.Infof("hello %s %d", "world", 42)
	if !strings.Contains(buf.String(), "[INFO] hello world 42") {
		t.Errorf("expected formatted output, got: %s", buf.String())
	}
}

func TestErrorf(t *testing.T) {
	l, buf := newTestLogger(t, "test")
	l.Errorf("code %d", 500)
	if !strings.Contains(buf.String(), "[ERROR] code 500") {
		t.Errorf("expected formatted output, got: %s", buf.String())
	}
}

func TestWarnf_Panics(t *testing.T) {
	l, _ := newTestLogger(t, "test")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Warnf to panic")
		}
	}()
	l.Warnf("warn %s", "formatted")
}

func TestFatal_Panics(t *testing.T) {
	l, buf := newTestLogger(t, "test")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Fatal to panic")
		}
		if !strings.Contains(buf.String(), "[FATAL] fatal message") {
			t.Errorf("expected [FATAL] in output, got: %s", buf.String())
		}
	}()
	l.Fatal("fatal message")
}

func TestFatalf_Panics(t *testing.T) {
	l, buf := newTestLogger(t, "test")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Fatalf to panic")
		}
		if !strings.Contains(buf.String(), "[FATAL] code 503") {
			t.Errorf("expected [FATAL] formatted output, got: %s", buf.String())
		}
	}()
	l.Fatalf("code %d", 503)
}

func TestDebugf_SuppressedByDefault(t *testing.T) {
	cfg := testConfig(t, "info")
	l, err := logs.NewLogger("test", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	buf := &bytes.Buffer{}
	l.AddWriter(buf)

	l.Debugf("hidden %d", 1)
	if strings.Contains(buf.String(), "hidden") {
		t.Error("Debugf should be suppressed when level is not debug")
	}
}

func TestDebugf_ShownWhenEnabled(t *testing.T) {
	cfg := testConfig(t, "debug")
	l, err := logs.NewLogger("test", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	buf := &bytes.Buffer{}
	l.AddWriter(buf)

	l.Debugf("value=%d", 7)
	if !strings.Contains(buf.String(), "[DEBUG] value=7") {
		t.Errorf("expected formatted debug output, got: %s", buf.String())
	}
}

func TestAddWriter_ReceivesOutput(t *testing.T) {
	l, _ := newTestLogger(t, "test")
	extra := &bytes.Buffer{}
	l.AddWriter(extra)
	l.Info("broadcast")
	if !strings.Contains(extra.String(), "[INFO] broadcast") {
		t.Errorf("extra writer did not receive output, got: %s", extra.String())
	}
}
