package utilities

import (
	"crypto/md5"
	"io/ioutil"
)

func GetMd5HashForFile(filepath string) ([16]byte, error) {

	fileContent, err := ioutil.ReadFile(filepath)
	if err != nil {
		return [16]byte{}, err
	}

	return md5.Sum(fileContent), nil
}
