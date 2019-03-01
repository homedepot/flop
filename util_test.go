package flop

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// memoizeTmpDir holds memoization info for the temporary directory
var memoizeTmpDir string

// unusedFileNum tracks a unique file identifier
var unusedFileNum int

// unusedDirNum tracks a unique dir identifier
var unusedDirNum int

// tmpDirPath gets the path of a temporary directory on the system
func tmpDirPath() string {
	// memoize
	if memoizeTmpDir != "" {
		return memoizeTmpDir
	}

	d, _ := ioutil.TempDir("", "")
	return d
}

// tmpDirPathUnused returns the path for a temp directory that does not exist yet
func tmpDirPathUnused() string {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	for {
		d = filepath.Join(d, fmt.Sprintf("%s%d", "dir", unusedDirNum))
		// we expect to see an error if the dir path is unused
		if _, err := os.Stat(d); err == nil {
			// bump file number if the file created with that number exists
			unusedDirNum += 1
		} else {
			return d
		}
	}
}

// tmpFile creates a new, empty temporary file and returns the full path
func tmpFile() string {
	src, err := ioutil.TempFile("", "*.txt")
	if err != nil {
		panic(fmt.Sprintf("temp file creation failed: %s", err))
	}
	defer func() {
		_ = src.Close()
	}()

	return src.Name()
}

// tmpFilePathUnused returns the path for a temp file that does not yet exist
func tmpFilePathUnused() string {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	// derive file name of potentially unused file
	tmpFile := func() string {
		return path.Join(d, fmt.Sprintf("%s%d", "file", unusedFileNum))
	}

	for {
		// we expect to see an error if the file path is unused
		if _, err := os.Stat(tmpFile()); err == nil {
			// bump file number if the file created with that number exists
			unusedFileNum += 1
		} else {
			return tmpFile()
		}
	}
}

func errContains(err error, substring string) bool {
	errString := fmt.Sprintf("%s", err)
	if strings.Contains(errString, substring) {
		return true
	}
	fmt.Println("error:", errString)
	fmt.Println("substring:", substring)
	return false
}

func debugLogger(msg string) {
	if debug {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Debug().Msg(msg)
	}
}

func infoLogger(msg string) {
	if debug {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Info().Msg(msg)
	}
}
