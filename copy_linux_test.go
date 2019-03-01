// +build linux

package flop

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestFileCopyOnDstWithInvalidPermissionsReturnsNoErrorWhenAtomic(t *testing.T) {
	assert := assert.New(t)
	// create and write to source inFile
	src := tmpFile()
	content := []byte("foo")
	assert.Nil(ioutil.WriteFile(src, content, 0644))

	dst := tmpFile()
	// explicitly set our dst inFile perms so that we cannot copy
	assert.Nil(os.Chmod(dst, 0040))

	assert.Nil(Copy(src, dst, Options{Atomic: true}))

	// make sure we can read out the correct content
	assert.Nil(os.Chmod(dst, 0655))
	b, err := ioutil.ReadFile(dst)
	assert.Nil(err)
	assert.Equal(content, b)

	// change perms back to ensure we can read to verify content
	assert.Nil(os.Chmod(dst, 0655))
}

func TestFileIsSymlink(t *testing.T) {
	assert := assert.New(t)
	old := tmpFile()
	new := tmpFilePathUnused()
	assert.Nil(os.Symlink(old, new))

	newFileInfo, err := os.Lstat(new)
	assert.Nil(err)
	f := File{
		Path:           new,
		fileInfoOnInit: newFileInfo,
	}
	assert.True(f.isSymlink())
}

func TestIsSymlinkFailsWithRegularFile(t *testing.T) {
	assert := assert.New(t)
	tmp := tmpFile()
	f := NewFile(tmp)
	f.setInfo()
	assert.False(f.isSymlink())
}

func TestPermissionsAfterCopy(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name             string
		atomic           bool
		dstShouldExist   bool
		srcPerms         os.FileMode
		expectedDstPerms os.FileMode
	}{
		{
			name:             "preserve_src_perms_when_dst_not_exist_0655",
			dstShouldExist:   false,
			srcPerms:         os.FileMode(0655),
			expectedDstPerms: os.FileMode(0655),
		},
		{
			name:             "preserve_src_perms_when_dst_not_exist_0777",
			dstShouldExist:   false,
			srcPerms:         os.FileMode(0777),
			expectedDstPerms: os.FileMode(0777),
		},
		{
			name:             "preserve_src_perms_when_dst_not_exist_0741",
			dstShouldExist:   false,
			srcPerms:         os.FileMode(0741),
			expectedDstPerms: os.FileMode(0741),
		},
		{
			name:             "preserve_dst_perms_when_dst_exists_0654",
			dstShouldExist:   true,
			srcPerms:         os.FileMode(0655),
			expectedDstPerms: os.FileMode(0654),
		},
		{
			name:             "preserve_dst_perms_when_dst_exists_0651",
			dstShouldExist:   true,
			srcPerms:         os.FileMode(0655),
			expectedDstPerms: os.FileMode(0651),
		},
		{
			name:             "preserve_dst_perms_when_dst_exists_0777",
			dstShouldExist:   true,
			srcPerms:         os.FileMode(0655),
			expectedDstPerms: os.FileMode(0777),
		},
		{
			name:             "preserve_dst_perms_when_dst_exists_0666",
			atomic:           false,
			dstShouldExist:   true,
			srcPerms:         os.FileMode(0655),
			expectedDstPerms: os.FileMode(0666),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := tmpFile()
			assert.Nil(os.Chmod(src, tt.srcPerms))
			var dst string
			if tt.dstShouldExist {
				dst = tmpFile()
				// set dst perms to ensure they are distinct beforehand
				assert.Nil(os.Chmod(dst, tt.expectedDstPerms))
			} else {
				dst = tmpFilePathUnused()
			}

			assert.Nil(Copy(src, dst, Options{
				Atomic:       tt.atomic,
				InfoLogFunc:  infoLogger,
				DebugLogFunc: debugLogger,
			}), "name: %s", tt.name)

			// check our perms
			d, err := os.Stat(dst)
			assert.Nil(err)
			dstPerms := d.Mode()
			assert.Equal(fmt.Sprint(tt.expectedDstPerms), fmt.Sprint(dstPerms))
		})
	}
}

func TestCopyingSymLinks(t *testing.T) {
	assert := assert.New(t)
	src := tmpFile()
	content := []byte("foo")
	assert.Nil(ioutil.WriteFile(src, content, 0655))
	srcSym := tmpFilePathUnused()
	assert.Nil(os.Symlink(src, srcSym))

	dstSym := tmpFilePathUnused()

	// copy sym link
	assert.Nil(SimpleCopy(srcSym, dstSym))

	// verify that dst is a sym link
	sfi, err := os.Lstat(dstSym)
	assert.Nil(err)
	assert.True(sfi.Mode()&os.ModeSymlink != 0)

	// verify content is the same in symlinked file
	b, err := ioutil.ReadFile(dstSym)
	assert.Nil(err)
	assert.Equal(content, b)
}

func TestCreatingHardLinksWithLinkOpt(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name string
		src  string
		dst  string
		opts Options
	}{
		{
			name: "absent_dst",
			src:  tmpFile(),
			dst:  tmpFilePathUnused(),
			opts: Options{Link: true},
		},
		//{  // TODO setup when force is implemented
		//	name: " existing_dst",
		//	src: tmpFile(),
		//	dst: tmpFile(),
		//	opts: Options{Link: true, Force: true},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := []byte("foo")
			assert.Nil(ioutil.WriteFile(tt.src, content, 0655))

			assert.Nil(Copy(tt.src, tt.dst, tt.opts))

			sFI, err := os.Stat(tt.src)
			assert.Nil(err)
			dFI, err := os.Stat(tt.dst)
			assert.Nil(err)
			assert.True(os.SameFile(sFI, dFI))
		})
	}
}
