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
			name:             "no_preserve_attrs_set_does_not_change_mode",
			srcPerms:         os.FileMode(0655),
			initDstPerms:     os.FileMode(0644),
			expectedDstPerms: os.FileMode(0655),
			options:          Options{},
		},
		{
			name:             "non_mode_preserve_attr_set_does_not_change_mode",
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
		attrs                     PreserveAttrs
		dstTimeShouldMatchSrcTime bool
	}{
		{
			name:                      "timestamps_attr_keeps_src_file_perms",
			attrs:                     PreserveAttrs{Timestamps: true},
			dstTimeShouldMatchSrcTime: true,
		},
		{
			name:                      "no_attrs_given_keeps_src_file_perms",
			dstTimeShouldMatchSrcTime: true,
		},
		{
			name:                      "non_mode_attr_set_does_not_preserve_timestamps",
			dstTimeShouldMatchSrcTime: false,
			attrs:                     PreserveAttrs{ownership: true},
		},
		{
			name:                      "none_attr_does_not_preserve_timestamps",
			dstTimeShouldMatchSrcTime: false,
			attrs:                     PreserveAttrs{None: true},
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
			opts := Options{
				InfoLogFunc:  infoLogger,
				DebugLogFunc: debugLogger,
				Preserve:     tt.attrs,
			}

			// copy
			assert.Nil(Copy(src, dst, opts), "failure on: %s", tt.name)

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

// TODO: we make too many assumptions, or take too many liberties, to properly assign and test ownership. revisit this.
//func TestPreserveOptionSetsDesiredOwnership(t *testing.T) {
//	assert := assert.New(t)
//	tests := []struct {
//		name                    string
//		attrs                 PreserveAttrs
//		dstOwnShouldMatchSrcOwn bool
//	}{
//		{
//			name:    "ownership_attr_keeps_ownership",
//			attrs: PreserveAttrs{ownership: true},
//			dstOwnShouldMatchSrcOwn: true,
//		},
//		//{
//		//	name:    "no_ownership_attr_does_not_keep_ownership",
//		//	attrs: PreserveAttrs{ownership: false},
//		//	dstOwnShouldMatchSrcOwn: false,
//		//},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			var src, dst string
//			{
//				src = tmpFile()
//				dst = tmpFile()
//				defer os.RemoveAll(src)
//				defer os.RemoveAll(dst)
//				err := os.Chown(dst, 111, 111)
//				assert.Nil(err, "unable to chown, err: '%s'", err)
//			}
//
//			// set default options
//			opts := Options{
//				InfoLogFunc:  infoLogger,
//				DebugLogFunc: debugLogger,
//				Preserve:     tt.attrs,
//			}
//
//			// copy
//			assert.Nil(Copy(src, dst, opts), "failure on: %s", tt.name)
//
//			// get ownership
//			var srcUid, dstUid, srcGid, dstGid uint32
//			{
//				srcFileInfo, err := os.Stat(src)
//				assert.Nil(err)
//				statT := srcFileInfo.Sys().(*syscall.Stat_t)
//				srcUid = statT.Uid
//				srcGid = statT.Gid
//
//				dstFileInfo, err := os.Stat(dst)
//				assert.Nil(err)
//				statT = dstFileInfo.Sys().(*syscall.Stat_t)
//				dstUid = statT.Uid
//				dstGid = statT.Gid
//			}
//
//			// check ownership
//			if tt.dstOwnShouldMatchSrcOwn {
//				assert.Equal(111, int(dstUid))
//				//assert.Equal(srcUid, dstUid)
//				//assert.Equal(srcGid, dstGid)
//			} else {
//				assert.NotEqual(srcUid, dstUid)
//				assert.NotEqual(srcGid, dstGid)
//			}
//		})
//	}
//}

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
