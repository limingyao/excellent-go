package csv

import (
	"bytes"
	"strings"
)

func New(header []string, lines [][]string) []byte {
	buffer := &bytes.Buffer{}
	buffer.WriteString("\xEF\xBB\xBF") // UTF-8 BOM
	buffer.WriteString(strings.Join(header, ","))
	buffer.WriteString("\n")
	for i := range lines {
		buffer.WriteString(strings.Join(lines[i], ","))
		buffer.WriteString("\n")
	}
	return buffer.Bytes()
}
