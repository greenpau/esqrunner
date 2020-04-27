package esqrunner

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func readFileBytes(fp string) ([]byte, error) {
	if fp[0] == '~' {
		hd, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		fp = filepath.Join(hd, fp[1:])
	}
	return ioutil.ReadFile(fp)
}
