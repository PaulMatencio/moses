package files

import (
	"bytes"
	"os"
)

func ReadBuffer(filename string) (*bytes.Buffer, error) {
	fp, e := os.Open(filename)
	if e == nil {
		defer fp.Close()
		fi, e := fp.Stat()
		var n int64
		n = fi.Size() + bytes.MinRead
		buf := make([]byte, 0, n)
		buffer := bytes.NewBuffer(buf)
		_, e = buffer.ReadFrom(fp)
		return buffer, e
	} else {
		return nil, e
	}
}
