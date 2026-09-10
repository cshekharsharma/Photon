package encoding

import (
	"bytes"
	"testing"
)

func TestIsValidUTF8(t *testing.T) {
	valid := []byte("hello✓")
	invalid := []byte{0xff, 0xfe, 0xfd}

	if !IsValidUTF8(valid) {
		t.Error("expected valid UTF-8")
	}
	if IsValidUTF8(invalid) {
		t.Error("expected invalid UTF-8")
	}
}

func TestFixUTF8(t *testing.T) {
	input := []byte("valid\xfftext")
	expected := "valid�text"
	if out := FixUTF8(input); out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestConvertToValidUTF8String(t *testing.T) {
	input := []byte("valid✓")
	out := ConvertToValidUTF8String(input)
	if out != "valid✓" {
		t.Errorf("expected valid UTF-8 string, got %q", out)
	}

	invalid := []byte("a\xffb")
	expected := "a�b"
	if out := ConvertToValidUTF8String(invalid); out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestNormalizeNFC(t *testing.T) {
	input := "é" // é as e + accent
	expected := "é"
	if out := NormalizeNFC(input); out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestNormalizeNFD(t *testing.T) {
	input := "é"
	expected := "é"
	if out := NormalizeNFD(input); out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestCountUnicodeCharacters(t *testing.T) {
	input := "a✓b"
	expected := 3
	if out := CountUnicodeCharacters(input); out != expected {
		t.Errorf("expected %d, got %d", expected, out)
	}
}

func TestStripInvalidUTF8(t *testing.T) {
	input := []byte("a\xffb")
	expected := []byte("ab")
	if out := StripInvalidUTF8(input); !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestConvertUTF8ToUTF16AndBack(t *testing.T) {
	original := "Hello✓"
	u16 := ConvertUTF8ToUTF16(original)
	back := ConvertUTF16ToUTF8(u16)
	if back != original {
		t.Errorf("expected %q, got %q", original, back)
	}
}

func TestConvertISO8859ToUTF8AndBack(t *testing.T) {
	original := "Hello!"
	iso, err := ConvertUTF8ToISO8859(original)
	if err != nil {
		t.Fatalf("ConvertUTF8ToISO8859 failed: %v", err)
	}
	utf8Str, err := ConvertISO8859ToUTF8(iso)
	if err != nil {
		t.Fatalf("ConvertISO8859ToUTF8 failed: %v", err)
	}
	if utf8Str != original {
		t.Errorf("expected %q, got %q", original, utf8Str)
	}
}

func TestDetectUTF16BOM(t *testing.T) {
	if DetectUTF16BOM([]byte{0xFF, 0xFE}) != "LE" {
		t.Error("expected LE BOM")
	}
	if DetectUTF16BOM([]byte{0xFE, 0xFF}) != "BE" {
		t.Error("expected BE BOM")
	}
	if DetectUTF16BOM([]byte{0x00, 0x00}) != "" {
		t.Error("expected no BOM")
	}
}

func TestConvertUTF16BytesToUTF8(t *testing.T) {
	utf16LE := []byte{0xFF, 0xFE, 'H', 0x00, 'i', 0x00} // BOM + "Hi"
	out, err := ConvertUTF16BytesToUTF8(utf16LE)
	if err != nil || out != "Hi" {
		t.Errorf("expected 'Hi', got %q (err: %v)", out, err)
	}

	utf16BE := []byte{0xFE, 0xFF, 0x00, 'H', 0x00, 'i'}
	out, err = ConvertUTF16BytesToUTF8(utf16BE)
	if err != nil || out != "Hi" {
		t.Errorf("expected 'Hi', got %q (err: %v)", out, err)
	}
}

func TestConvertUTF16BytesToUTF8_Errors(t *testing.T) {
	_, err := ConvertUTF16BytesToUTF8([]byte{0xFF})
	if err == nil {
		t.Error("expected error for odd-length input")
	}
	_, err = ConvertUTF16BytesToUTF8([]byte{0x00, 0x00})
	if err == nil {
		t.Error("expected error for missing BOM")
	}
}
