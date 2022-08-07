package utilities

import (
	"io/ioutil"
	"log"
)

func CopyFile(sourceFile, targetFile string) {
	//Read all the contents of the  original file
	bytesRead, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Println(err)
	}

	//Copy all the contents to the desitination file
	err = ioutil.WriteFile(targetFile, bytesRead, 0755)
	if err != nil {
		log.Println(err)
	}
}
