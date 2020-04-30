package esqrunner

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func expandHomePath(fp string) (string, error) {
	if fp[0] != '~' {
		return fp, nil
	}
	hd, err := os.UserHomeDir()
	if err != nil {
		return fp, err
	}
	fp = filepath.Join(hd, fp[1:])
	return fp, nil
}

func readFileBytes(fp string) ([]byte, error) {
	var err error
	fp, err = expandHomePath(fp)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(fp)
}

func writeToFile(fp string, data string) error {
	return ioutil.WriteFile(fp, []byte(data), 0644)
}
