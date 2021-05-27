package files

import (
	goLog "github.com/paulmatencio/moses/user/goLog"
	"os"
)

func MakeDir(dirname string, mode os.FileMode) error {
	var err error
	if !Exist(dirname) {
		goLog.Info.Println("Marking Dir>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>:", dirname)
		if err = os.MkdirAll(dirname, mode); err != nil {
			goLog.Error.Println("Making Directory", dirname, err)
		}
	} else {
		err = nil
	}
	return err
}
