package contracts

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

func GeneratePDF(title, body string) ([]byte, error) {
	var buf bytes.Buffer
	offsets := []int{}

	write := func(s string) {
		offsets = append(offsets, buf.Len())
		buf.WriteString(s)
	}

	buf.WriteString("%PDF-1.4\n")

	write("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	write("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	write("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	lines := strings.Split(body, "\n")
	var stream bytes.Buffer
	stream.WriteString("BT\n/F1 12 Tf\n")
	stream.WriteString(fmt.Sprintf("50 780 Td\n/F1 16 Tf\n(%s) Tj\n", pdfEscape(title)))
	stream.WriteString(fmt.Sprintf("/F1 10 Tf\n0 -20 Td\n(Generated: %s) Tj\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC")))
	stream.WriteString("/F1 11 Tf\n0 -25 Td\n")
	for _, line := range lines {
		stream.WriteString(fmt.Sprintf("0 -14 Td\n(%s) Tj\n", pdfEscape(line)))
	}
	stream.WriteString("ET\n")
	streamBytes := stream.Bytes()

	write(fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(streamBytes), streamBytes))

	write("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 5 0 R /Resources << /Font << /F1 4 0 R >> >> >>\nendobj\n")

	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString(fmt.Sprintf("0 %d\n", len(offsets)+1))
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}

	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", len(offsets)+1))
	buf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefOffset))

	return buf.Bytes(), nil
}

func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}
