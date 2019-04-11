// +build linux darwin

package flop

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"
)

// TODO: it's difficult to create paths that we do not have access to. revisit this.
//func TestFileExistsReturnsFalseWhenPermissionDenied(t *testing.T) {
//	assert := assert.New(t)
//	tmpDir, err := ioutil.TempDir("", "")
//	assert.Nil(err)
//	unreachableDir := filepath.Join(tmpDir, "unreachable")
//	unreachableFile := filepath.Join(unreachableDir, "file.txt")
//	os.Mkdir(unreachableDir, 0777)
//	ioutil.WriteFile(unreachableFile, []byte("foo"), 0777)
//	assert.Nil(os.Chown(unreachableDir, 0, 0))
//	assert.Nil(os.Chmod(unreachableDir, 0700))
//
//	u, err := user.Current()
//	assert.Nil(err)
//	assert.NotEqual("0", u.Uid, "it looks like you are running as UID 0, which breaks tests designed to find permission errors.. don't do that please")
//
//	f, err := ioutil.ReadDir(tmpDir)
//	assert.Nil(err)
//	fmt.Println(f[0])
//}

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
		Path:     new,
		fileInfo: &newFileInfo,
	}
	assert.True(f.isSymlink())
}

func TestIsSymlinkFailsWithRegularFile(t *testing.T) {
	assert := assert.New(t)
	tmp := tmpFile()
	f := NewFile(tmp)
	assert.False(f.isSymlink())
}

//func TestPermissionsAfterCopyWithoutPreserveOptions(t *testing.T) {  // TODO: reuse these with preserve defaults or delete
//	assert := assert.New(t)
//	tests := []struct {
//		name string
//		// if dst should exist we'll create it and assign expectedDstPerms to it
//		dstShouldExist   bool
//		srcPerms         os.FileMode
//		expectedDstPerms os.FileMode
//		options          Options
//	}{
//		{
//			name:             "dst_not_exist_0655",
//			dstShouldExist:   false,
//			srcPerms:         os.FileMode(0655),
//			expectedDstPerms: os.FileMode(0655),
//		},
//		{
//			name:             "dst_not_exist_0777",
//			dstShouldExist:   false,
//			srcPerms:         os.FileMode(0777),
//			expectedDstPerms: os.FileMode(0777),
//		},
//		{
//			name:             "preserve_dst_perms_when_dst_exists_0654",
//			dstShouldExist:   true,
//			srcPerms:         os.FileMode(0655),
//			expectedDstPerms: os.FileMode(0654),
//		},
//		{
//			name:             "preserve_dst_perms_when_dst_exists_0651",
//			dstShouldExist:   true,
//			srcPerms:         os.FileMode(0655),
//			expectedDstPerms: os.FileMode(0651),
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			src := tmpFile()
//			assert.Nil(os.Chmod(src, tt.srcPerms))
//			var dst string
//			if tt.dstShouldExist {
//				dst = tmpFile()
//				// set dst perms to ensure they are distinct beforehand
//				assert.Nil(os.Chmod(dst, tt.expectedDstPerms))
//			} else {
//				dst = tmpFilePathUnused()
//			}
//
//			// set default options
//			tt.options.InfoLogFunc = infoLogger
//			tt.options.DebugLogFunc = debugLogger
//
//			// copy
//			assert.Nil(Copy(src, dst, tt.options), "failure on: %s", tt.name)
//
//			// check our perms
//			d, err := os.Stat(dst)
//			assert.Nil(err)
//			dstPerms := d.Mode()
//			assert.Equal(fmt.Sprint(tt.expectedDstPerms), fmt.Sprint(dstPerms))
//		})
//	}
//}

func TestPreserveOptionsSetsDesiredPermissions(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name             string
		srcPerms         os.FileMode
		initDstPerms     os.FileMode
		expectedDstPerms os.FileMode
		options          Options
	}{
		{
			name:             "preserve_mode_keeps_src_file_perms",
			srcPerms:         os.FileMode(0655),
			initDstPerms:     os.FileMode(0644),
			expectedDstPerms: os.FileMode(0655),
			options:          Options{Preserve: PreserveAttrs{Mode: true}},
		},
		{
			name:             "preserve_mode_with_no_mode_set",
			srcPerms:         os.FileMode(0655),
			initDstPerms:     os.FileMode(0644),
			expectedDstPerms: os.FileMode(0655),
			options:          Options{},
		},
		{
			name:             "do_not_preserve_mode_when_mode_is_not_set",
			srcPerms:         os.FileMode(0655),
			initDstPerms:     os.FileMode(0641),
			expectedDstPerms: os.FileMode(0641),
			options:          Options{Preserve: PreserveAttrs{Timestamps: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var src, dst string
			{
				src = tmpFile()
				assert.Nil(os.Chmod(src, tt.srcPerms))
				dst = tmpFile()
				assert.Nil(os.Chmod(dst, tt.initDstPerms))
			}

			// set default options
			tt.options.InfoLogFunc = infoLogger
			tt.options.DebugLogFunc = debugLogger

			// copy
			assert.Nil(Copy(src, dst, tt.options), "failure on: %s", tt.name)

			// check the perms
			d, err := os.Stat(dst)
			assert.Nil(err)
			dstPerms := d.Mode()
			assert.Equal(fmt.Sprint(tt.expectedDstPerms), fmt.Sprint(dstPerms), "subtest: %s", tt.name)
		})
	}
}

func TestPreserveOptionSetsDesiredTimestamps(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name                      string
		options                   Options
		dstTimeShouldMatchSrcTime bool
	}{
		{
			name:                      "preserve_mode_keeps_src_file_perms",
			options:                   Options{Preserve: PreserveAttrs{Timestamps: true}},
			dstTimeShouldMatchSrcTime: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var src, dst string
			{
				d := time.Date(2000, time.January, 5, 1, 2, 3, 4, time.Local)
				src = tmpFile()
				dst = tmpFile()
				defer os.RemoveAll(src)
				defer os.RemoveAll(dst)
				assert.Nil(os.Chtimes(dst, d, d))
			}

			// set default options
			tt.options.InfoLogFunc = infoLogger
			tt.options.DebugLogFunc = debugLogger

			// copy
			assert.Nil(Copy(src, dst, tt.options), "failure on: %s", tt.name)

			var srcATime, dstATime time.Time // access times
			var srcMTime, dstMTime time.Time // mod times
			{
				// get src times
				srcFileInfo, err := os.Stat(src)
				assert.Nil(err)
				srcStatT := srcFileInfo.Sys().(*syscall.Stat_t)
				srcATime = time.Unix(srcStatT.Atim.Sec, srcStatT.Atim.Nsec)
				srcMTime = time.Unix(srcStatT.Mtim.Sec, srcStatT.Mtim.Nsec)

				// get dst times
				dstFileInfo, err := os.Stat(dst)
				assert.Nil(err)
				dstStatT := dstFileInfo.Sys().(*syscall.Stat_t)
				dstATime = time.Unix(dstStatT.Atim.Sec, dstStatT.Atim.Nsec)
				dstMTime = time.Unix(dstStatT.Mtim.Sec, dstStatT.Mtim.Nsec)
			}

			// check the times
			if tt.dstTimeShouldMatchSrcTime {
				assert.Equal(srcATime, dstATime)
				assert.Equal(srcMTime, dstMTime)
			} else {
				assert.NotEqual(srcATime, dstATime)
				assert.NotEqual(srcMTime, dstMTime)
			}
		})
	}
}

func TestCopyingSymLinks(t *testing.T) {
	assert := assert.New(t)
	src := tmpFile()
	defer os.RemoveAll(src)
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
			defer os.RemoveAll(tt.src)
			defer os.RemoveAll(tt.dst)

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
