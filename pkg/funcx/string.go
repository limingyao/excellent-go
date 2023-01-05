package funcx

import (
	"bytes"
	"fmt"
)

func JoinToString[T any](args []T, sep string) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(fmt.Sprintf("%v", v))
	}

	return buf.String()
}
