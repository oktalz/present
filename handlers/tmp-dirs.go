package handlers

import (
	"errors"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	tmpDirs      = map[string]string{}
	tmpDirsMutex = &sync.Mutex{}
)

func GetWorkingTmpDir(workingDir string) (string, error) {
	tmpDirsMutex.Lock()
	defer tmpDirsMutex.Unlock()
	createAFixedTmpDIr := false
	additionalDir := ""
	if strings.HasPrefix(workingDir, "mkdir{") {
		createAFixedTmpDIr = true
		workingDir = strings.TrimPrefix(workingDir, "mkdir{")
		// additionalDir is everything after first }
		index := strings.Index(workingDir, "}")
		if index != -1 {
			additionalDir = workingDir[index+1:]
			workingDir = workingDir[:index]
		}
	}
	if after, ok := strings.CutPrefix(workingDir, "{"); ok {
		workingDir = after
		// additionalDir is everything after first }
		index := strings.Index(workingDir, "}")
		if index != -1 {
			additionalDir = workingDir[index+1:]
			workingDir = workingDir[:index]
		}
		tmpDir, ok := tmpDirs[workingDir]
		if ok {
			return path.Join(tmpDir, additionalDir), nil
		}
		return "", errors.New("tmp dir " + workingDir + " not found")
	}

	tmpDir := os.TempDir() + "/present-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	err := os.MkdirAll(path.Join(tmpDir, additionalDir), 0o755)
	if err != nil {
		return "", err
	}

	if createAFixedTmpDIr {
		tmpDirs[workingDir] = tmpDir
	}

	return path.Join(tmpDir, additionalDir), nil
}
