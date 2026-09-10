// utf_utils.go
package encoding

import (
	"bytes"
	"errors"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/unicode/norm"
)

// IsValidUTF8 returns true if the input byte slice is valid UTF-8.
func IsValidUTF8(data []byte) bool {
	return utf8.Valid(data)
}

// FixUTF8 replaces invalid UTF-8 byte sequences in the input with the Unicode replacement character (\uFFFD).
func FixUTF8(data []byte) string {
	var buf bytes.Buffer
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			buf.WriteRune('\uFFFD')
			data = data[1:]
		} else {
			buf.WriteRune(r)
			data = data[size:]
		}
	}
	return buf.String()
}

// ConvertToValidUTF8String converts a potentially invalid UTF-8 byte slice into a valid UTF-8 string.
// It returns the original string if valid, or a fixed version with invalid bytes replaced.
func ConvertToValidUTF8String(data []byte) string {
	if IsValidUTF8(data) {
		return string(data)
	}
	return FixUTF8(data)
}

// NormalizeNFC returns the NFC (Normalization Form C) version of the input string.
// Useful for canonical representation and string comparisons.
func NormalizeNFC(s string) string {
	return norm.NFC.String(s)
}

// NormalizeNFD returns the NFD (Normalization Form D) version of the input string.
// This form decomposes characters into base characters and combining marks.
func NormalizeNFD(s string) string {
	return norm.NFD.String(s)
}

// CountUnicodeCharacters returns the number of Unicode code points (runes) in the input string.
func CountUnicodeCharacters(s string) int {
	return utf8.RuneCountInString(s)
}

// StripInvalidUTF8 removes any invalid UTF-8 sequences from the input byte slice.
// Valid sequences are preserved; invalid ones are skipped entirely.
func StripInvalidUTF8(data []byte) []byte {
	var buf bytes.Buffer
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			data = data[1:]
		} else {
			buf.WriteRune(r)
			data = data[size:]
		}
	}
	return buf.Bytes()
}

// ConvertUTF8ToUTF16 converts a UTF-8 string to a slice of UTF-16 code units (LE format).
func ConvertUTF8ToUTF16(s string) []uint16 {
	return utf16.Encode([]rune(s))
}

// ConvertUTF16ToUTF8 converts a slice of UTF-16 code units to a UTF-8 encoded string.
func ConvertUTF16ToUTF8(u16 []uint16) string {
	return string(utf16.Decode(u16))
}

// ConvertISO8859ToUTF8 decodes an ISO-8859-1 (Latin-1) byte slice into a UTF-8 string.
func ConvertISO8859ToUTF8(input []byte) (string, error) {
	return charmap.ISO8859_1.NewDecoder().String(string(input))
}

// ConvertUTF8ToISO8859 encodes a UTF-8 string into ISO-8859-1 (Latin-1) format bytes.
func ConvertUTF8ToISO8859(input string) ([]byte, error) {
	return charmap.ISO8859_1.NewEncoder().Bytes([]byte(input))
}

// DetectUTF16BOM checks for a UTF-16 byte order mark (BOM) at the start of the byte slice.
// Returns "LE" for little-endian, "BE" for big-endian, or an empty string if no BOM is found.
func DetectUTF16BOM(data []byte) string {
	if len(data) >= 2 {
		switch {
		case data[0] == 0xFF && data[1] == 0xFE:
			return "LE"
		case data[0] == 0xFE && data[1] == 0xFF:
			return "BE"
		}
	}
	return ""
}

// ConvertUTF16BytesToUTF8 converts a byte slice encoded in UTF-16 (with BOM) into a UTF-8 string.
// Supports both little-endian and big-endian formats.
func ConvertUTF16BytesToUTF8(data []byte) (string, error) {
	if len(data)%2 != 0 {
		return "", errors.New("invalid UTF-16 byte length")
	}

	swap := false
	switch DetectUTF16BOM(data) {
	case "LE":
		data = data[2:] // strip BOM
	case "BE":
		data = data[2:] // strip BOM
		swap = true
	default:
		return "", errors.New("missing UTF-16 BOM")
	}

	u16 := make([]uint16, len(data)/2)
	for i := 0; i < len(u16); i++ {
		if swap {
			u16[i] = uint16(data[2*i+1]) | uint16(data[2*i])<<8
		} else {
			u16[i] = uint16(data[2*i]) | uint16(data[2*i+1])<<8
		}
	}

	return ConvertUTF16ToUTF8(u16), nil
}
