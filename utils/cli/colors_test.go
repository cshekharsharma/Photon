package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

func TestSetFg(t *testing.T) {
	SetFg(Red)
	if GetFgColor() != Red {
		t.Errorf("Expected %v, got %v", Red, GetFgColor())
	}

	SetFg(BoldRed)
	if GetFgColor() != BoldRed || !IsBold() {
		t.Errorf("Expected %v and bold, got %v and bold=%v", BoldRed, GetFgColor(), IsBold())
	}
}

func TestSetBg(t *testing.T) {
	SetBg(Blue)
	if GetBgColor() != Blue {
		t.Errorf("Expected %v, got %v", Blue, GetBgColor())
	}
}

func TestSetBold(t *testing.T) {
	SetBold(true)
	if !IsBold() {
		t.Errorf("Expected bold to be true, got %v", IsBold())
	}

	SetBold(false)
	if IsBold() {
		t.Errorf("Expected bold to be false, got %v", IsBold())
	}
}

func TestSetUnderline(t *testing.T) {
	SetUnderline(true)
	if !IsUnderline() {
		t.Errorf("Expected underline to be true, got %v", IsUnderline())
	}

	SetUnderline(false)
	if IsUnderline() {
		t.Errorf("Expected underline to be false, got %v", IsUnderline())
	}
}

func TestSetReversed(t *testing.T) {
	SetReversed(true)
	if !IsReversed() {
		t.Errorf("Expected reversed to be true, got %v", IsReversed())
	}

	SetReversed(false)
	if IsReversed() {
		t.Errorf("Expected reversed to be false, got %v", IsReversed())
	}
}

func TestSetAll(t *testing.T) {
	SetAll(Red, Blue, true, true, true)
	if GetFgColor() != Red || GetBgColor() != Blue || !IsBold() || !IsUnderline() || !IsReversed() {
		t.Errorf("Expected fg=%v, bg=%v, bold=%v, underline=%v, reversed=%v, got fg=%v, bg=%v, bold=%v, underline=%v, reversed=%v",
			Red, Blue, true, true, true, GetFgColor(), GetBgColor(), IsBold(), IsUnderline(), IsReversed())
	}
}

func TestResetAll(t *testing.T) {
	ResetAll()
	if GetFgColor() != Default || GetBgColor() != Default || IsBold() || IsUnderline() || IsReversed() {
		t.Errorf("Expected reset to default, got fg=%v, bg=%v, bold=%v, underline=%v, reversed=%v",
			GetFgColor(), GetBgColor(), IsBold(), IsUnderline(), IsReversed())
	}
}

func TestPrint(t *testing.T) {
	output := captureOutput(t, func() {
		Print(Green, "Hello, world!")
	})
	expected := "\033[0;32mHello, world!\033[0;39;49;22;24;27m"
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}

	output = captureOutput(t, func() {
		Print(BoldGreen, "Hello, world!")
	})
	expected = "\033[0;32m\033[1mHello, world!\033[0;39;49;22;24;27m"
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}
}

func TestPrintln(t *testing.T) {
	output := captureOutput(t, func() {
		Println(Green, "Hello, world!")
	})
	expected := "\033[0;32mHello, world!\n\033[0;39;49;22;24;27m"
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}

	output = captureOutput(t, func() {
		Println(BoldGreen, "Hello, world!")
	})
	expected = "\033[0;32m\033[1mHello, world!\n\033[0;39;49;22;24;27m"
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}
}

func TestPrintf(t *testing.T) {
	output := captureOutput(t, func() {
		Printf(Green, "Hello, %s!", "world")
	})
	expected := "\033[0;32mHello, world!\033[0;39;49;22;24;27m"
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}

	output = captureOutput(t, func() {
		Printf(BoldGreen, "Hello, %s!", "world")
	})
	expected = "\033[0;32m\033[1mHello, world!\033[0;39;49;22;24;27m"
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}
}

func TestFprint(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf, Green, "Hello, world!")
	expected := "\033[0;32mHello, world!\033[0;39;49;22;24;27m"
	if buf.String() != expected {
		t.Errorf("Expected %v, got %v", expected, buf.String())
	}

	buf.Reset()
	Fprint(&buf, BoldGreen, "Hello, world!")
	expected = "\033[0;32m\033[1mHello, world!\033[0;39;49;22;24;27m"
	if buf.String() != expected {
		t.Errorf("Expected %v, got %v", expected, buf.String())
	}
}

func TestFprintln(t *testing.T) {
	var buf bytes.Buffer
	Fprintln(&buf, Green, "Hello, world!")
	expected := "\033[0;32mHello, world!\n\033[0;39;49;22;24;27m"
	if buf.String() != expected {
		t.Errorf("Expected %v, got %v", expected, buf.String())
	}

	buf.Reset()
	Fprintln(&buf, BoldGreen, "Hello, world!")
	expected = "\033[0;32m\033[1mHello, world!\n\033[0;39;49;22;24;27m"
	if buf.String() != expected {
		t.Errorf("Expected %v, got %v", expected, buf.String())
	}
}

func TestFprintf(t *testing.T) {
	var buf bytes.Buffer
	Fprintf(&buf, Green, "Hello, %s!", "world")
	expected := "\033[0;32mHello, world!\033[0;39;49;22;24;27m"
	if buf.String() != expected {
		t.Errorf("Expected %v, got %v", expected, buf.String())
	}

	buf.Reset()
	Fprintf(&buf, BoldGreen, "Hello, %s!", "world")
	expected = "\033[0;32m\033[1mHello, world!\033[0;39;49;22;24;27m"
	if buf.String() != expected {
		t.Errorf("Expected %v, got %v", expected, buf.String())
	}
}

type failingWriter struct {
	failAt int
	writes int
}

func (w *failingWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}

func TestApplySettings_WriteError(t *testing.T) {
	applySettings(&failingWriter{failAt: 1})
}

func TestResetLine_WriteError(t *testing.T) {
	resetLine(&failingWriter{failAt: 1})
}

func TestWriteWithColor_Errors(t *testing.T) {
	t.Run("ColorWriteError", func(t *testing.T) {
		Fprint(&failingWriter{failAt: 1}, Green, "hello")
	})

	t.Run("BoldWriteError", func(t *testing.T) {
		Fprint(&failingWriter{failAt: 2}, BoldGreen, "hello")
	})

	t.Run("ContentWriteError", func(t *testing.T) {
		Fprint(&failingWriter{failAt: 2}, Green, "hello")
	})

	t.Run("ResetWriteError", func(t *testing.T) {
		Fprint(&failingWriter{failAt: 3}, Green, "hello")
	})
}

// Helper function to capture output.
func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	stdout := os.Stdout
	os.Stdout = w

	f()
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to copy captured output: %v", err)
	}
	os.Stdout = stdout
	return buf.String()
}
