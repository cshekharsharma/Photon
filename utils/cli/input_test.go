package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func simulateInput(t *testing.T, input string) func() {
	t.Helper()
	saved := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	if _, err := w.WriteString(input + "\n"); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close input writer: %v", err)
	}
	os.Stdin = r
	return func() { os.Stdin = saved }
}

func TestPrompt(t *testing.T) {
	restore := simulateInput(t, "gopher")
	defer restore()
	result := Prompt("Enter name:")
	if result != "gopher" {
		t.Errorf("expected 'gopher', got '%s'", result)
	}
}

func TestPromptWithDefault(t *testing.T) {
	restore := simulateInput(t, "")
	defer restore()
	result := PromptWithDefault("Enter city", "Bangalore")
	if result != "Bangalore" {
		t.Errorf("expected default 'Bangalore', got '%s'", result)
	}

	restore = simulateInput(t, "Gurgaon")
	defer restore()
	result = PromptWithDefault("Enter city", "Bangalore")
	if result != "Gurgaon" {
		t.Errorf("expected default 'Bangalore', got '%s'", result)
	}
}

func TestConfirmYes(t *testing.T) {
	restore := simulateInput(t, "y")
	defer restore()
	if !Confirm("Continue?") {
		t.Errorf("expected confirmation true for 'y'")
	}
}

func TestConfirmNo(t *testing.T) {
	restore := simulateInput(t, "no")
	defer restore()
	if Confirm("Proceed?") {
		t.Errorf("expected confirmation false for 'no'")
	}
}

func TestPromptPassword(t *testing.T) {
	passwordReader = func(fd int) ([]byte, error) {
		return []byte("supersecret"), nil
	}

	output := captureStdout(t, func() {
		result := PromptPassword("Enter password:")
		if result != "supersecret" {
			t.Errorf("expected 'supersecret', got '%s'", result)
		}
	})

	if !strings.Contains(output, "Enter password:") {
		t.Errorf("expected prompt to contain 'Enter password:', got '%s'", output)
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	var buf bytes.Buffer
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan struct{})
	go func() {
		if _, err := io.Copy(&buf, r); err != nil {
			t.Errorf("failed to copy stdout: %v", err)
		}
		close(done)
	}()

	f()
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close stdout writer: %v", err)
	}
	<-done
	os.Stdout = old
	return buf.String()
}
