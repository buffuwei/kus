package tools

import (
	"os"
	"regexp"
	"time"

	"go.uber.org/zap"
)

var exp = regexp.MustCompile(`.*\.log`)

func Clean(dir string) {

	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		zap.S().Errorf("read dir %s failed, err: %v \n", dir, err)
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			Clean(dir + "/" + fileInfo.Name())
		} else {
			filePath := dir + "/" + fileInfo.Name()
			fi, _ := os.Stat(filePath)

			if fi.ModTime().Before(time.Now().AddDate(0, 0, -2)) &&
				exp.MatchString(fileInfo.Name()) {
				os.Remove(filePath)
				zap.S().Infof("remove file %s success \n", filePath)
			}
		}
	}

}

func ListDirWithoutSubdirs(dirPth string) ([]string, error) {
	fileInfos, err := os.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			files = append(files, fileInfo.Name())
		} else {

		}
	}
	return files, nil
}
