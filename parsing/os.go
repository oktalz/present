package parsing

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func getOSPath(filePath string) string {
	filePath = strings.ReplaceAll(filePath, "/", string(os.PathSeparator))
	filePath = strings.ReplaceAll(filePath, "\\", string(os.PathSeparator))

	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	filePath = path.Join(wd, filePath)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return ""
	}

	if fileInfo.IsDir() {
		return filePath
	}
	fmt.Println(filePath, "is not a directory")
	return filePath
}
