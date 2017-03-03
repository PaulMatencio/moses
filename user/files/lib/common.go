package files

import (
	"time"
)

var Timeout = time.Duration(100)

type Responses struct {
	Err  error
	Body []byte
}
