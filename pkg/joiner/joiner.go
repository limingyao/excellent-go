package joiner

import (
	"bytes"
	"fmt"
)

func Floats32(separator string, args ...float32) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%f", v))
	}

	return buf.String()
}

func Floats64(separator string, args ...float64) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%f", v))
	}

	return buf.String()
}

func Ints8(separator string, args ...int8) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%d", v))
	}

	return buf.String()
}

func Uints8(separator string, args ...uint8) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%d", v))
	}

	return buf.String()
}

func Ints(separator string, args ...int) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%d", v))
	}

	return buf.String()
}

func Uints(separator string, args ...uint) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%d", v))
	}

	return buf.String()
}

func Ints32(separator string, args ...int32) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%d", v))
	}

	return buf.String()
}

func Uints32(separator string, args ...uint32) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%d", v))
	}

	return buf.String()
}

func Ints64(separator string, args ...int64) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%d", v))
	}

	return buf.String()
}

func Uints64(separator string, args ...uint64) string {
	var buf bytes.Buffer

	for _, v := range args {
		if buf.Len() != 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(fmt.Sprintf("%d", v))
	}

	return buf.String()
}
