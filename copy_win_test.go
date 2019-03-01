// +build windows

package flop

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestForPresenceOfFileCopyErrors(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name                 string
		copyFunc             func(string, os.FileInfo, string, os.FileInfo) error
		inFile               string
		outFile              string
		errSubstringExpected string
	}{
		{
			name:                 "src_path_which_cannot_be_opened",
			inFile:               "////",
			outFile:              tmpFile(),
			errSubstringExpected: ErrCannotStatFile.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SimpleCopy(tt.inFile, tt.outFile)
			assert.True(errContains(err, tt.errSubstringExpected))
		})
	}
}

func TestFileCopyOnDstWithInvalidPermissionsReturnsAccessError(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name         string
		opts         Options
		errSubstring string
	}{
		{"atomic", Options{Atomic: true}, ErrCannotRenameTempFile.Error()},
		{"atomic", Options{Atomic: false}, ErrCannotOpenOrCreateDstFile.Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create and write to source inFile
			src := tmpFile()
			content := []byte("foo")
			assert.Nil(ioutil.WriteFile(src, content, 0644))

			dst := tmpFile()
			// explicitly set our dst inFile perms so that we cannot copy
			assert.Nil(os.Chmod(dst, 0111))

			err := Copy(src, dst, tt.opts)
			assert.True(errContains(err, tt.errSubstring), "err is: %s", err)

			// change perms back to ensure we can read to verify content
			assert.Nil(os.Chmod(dst, 0655))
		})
	}
}
