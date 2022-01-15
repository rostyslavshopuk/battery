package battery

import (
	"io/fs"
	"io/ioutil"
)

// IOUtil is used to perform IO operations, we define it here so that we can mock it during the tests
type IOUtil interface {
	ReadFile(fileName string) ([]byte, error)
	ReadDir(dirname string) ([]fs.FileInfo, error)
}

type CustomIOUtil struct{}

func (c CustomIOUtil) ReadDir(dirname string) ([]fs.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (c CustomIOUtil) ReadFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fileName)
}

var MyIOUtil IOUtil = CustomIOUtil{}
