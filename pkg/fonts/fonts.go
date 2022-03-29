package fonts

import (
	"bytes"
	_ "embed"
	"io"
)

//go:embed 微软雅黑.ttf
var msyhFont []byte

// MsyhReader 微软雅黑体 reader
func MsyhReader() io.Reader {
	return bytes.NewReader(msyhFont)
}
