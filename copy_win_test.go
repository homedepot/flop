// +build windows

package flop

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

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

// ensure that we are properly cleaning file paths before use
func TestCleaningWindowsFilePaths(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"src_trailing_back_slashes", tmpFile() + "\\\\\\\\", tmpFile()},
		{"dst_trailing_back_slashes", tmpFile(), tmpFile() + "\\\\\\\\"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(SimpleCopy(tt.in, tt.out))
		})
	}
}
