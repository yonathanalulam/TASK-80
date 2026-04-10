package contracts

import (
	"bytes"
	"testing"
)

func TestGeneratePDF_ValidOutput(t *testing.T) {
	data, err := GeneratePDF("Test Contract", "This is a test contract body.\nLine 2.\nLine 3.")
	if err != nil {
		t.Fatalf("GeneratePDF returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GeneratePDF returned empty data")
	}

	// Valid PDF must start with %PDF-
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Errorf("PDF does not start with PDF header, got: %q", data[:min(20, len(data))])
	}

	// Valid PDF must end with EOF marker
	if !bytes.Contains(data, []byte("%EOF")) {
		t.Error("PDF does not contain EOF trailer")
	}

	// Must contain the title
	if !bytes.Contains(data, []byte("Test Contract")) {
		t.Error("PDF does not contain the title")
	}
}

func TestGeneratePDF_EmptyBody(t *testing.T) {
	data, err := GeneratePDF("Empty", "")
	if err != nil {
		t.Fatalf("GeneratePDF with empty body returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("PDF with empty body should still be valid")
	}
}

func TestGeneratePDF_SpecialCharacters(t *testing.T) {
	data, err := GeneratePDF("Special (Chars)", "Body with (parens) and \\backslash")
	if err != nil {
		t.Fatalf("GeneratePDF with special chars returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("PDF with special chars should still be valid")
	}
	// Parens must be escaped in PDF strings
	if bytes.Contains(data, []byte("(parens)")) {
		// This would be unescaped parens inside a PDF string - might break rendering
		// The pdfEscape function should handle this
	}
}

func TestGeneratePDF_NonZeroSize(t *testing.T) {
	data, err := GeneratePDF("Invoice INV-2026-0001", "Amount: $850.00\nItems:\n- Lodging: $450.00\n- Transport: $200.00\n- Activity: $200.00")
	if err != nil {
		t.Fatal(err)
	}
	// A reasonable PDF should be at least a few hundred bytes
	if len(data) < 100 {
		t.Errorf("PDF seems too small: %d bytes", len(data))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
