// files project files.go
package files

import (
	"os"
)

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Exist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false

}
