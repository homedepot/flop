package flop

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// debug will perform advanced logging if set to true
// set to false to keep test results more terse
const debug = false

func TestFileExists(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name        string
		filePath    string
		shouldExist bool
	}{
		{"exists", tmpFile(), true},
		{"not_exists", tmpFilePathUnused(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFile(tt.filePath)
			assert.Equal(tt.shouldExist, f.Exists())
		})
	}
}

func TestFileContentInCopy(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name    string
		inFile  string
		outFile string
	}{
		{name: "basic_file_copy", inFile: tmpFile(), outFile: tmpFile()},
		{name: "unused_dst_file_path", inFile: tmpFile(), outFile: tmpFilePathUnused()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := []byte("foo")

			err := ioutil.WriteFile(tt.inFile, content, 0655)
			assert.Nil(err)

			err = SimpleCopy(tt.inFile, tt.outFile)
			assert.Nil(err, "err is:", err)

			outFileContent, err := ioutil.ReadFile(tt.outFile)
			assert.Nil(err)
			assert.Equal(content, outFileContent)
		})
	}
}

func TestErrorsInCopy(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name                 string
		inFile               string
		outFile              string
		options              Options
		errExpected          bool
		errSubstringExpected string
	}{
		{
			name:                 "file_does_not_exist",
			inFile:               tmpFilePathUnused(),
			outFile:              tmpFile(),
			errExpected:          true,
			errSubstringExpected: ErrFileNotExist.Error(),
		},
		{
			name:        "file_exist",
			inFile:      tmpFile(),
			outFile:     tmpFile(),
			errExpected: false,
		},
		{
			name:                 "dst_path_which_cannot_be_opened_or_created",
			inFile:               tmpFile(),
			outFile:              "/path/that/is/inaccessible",
			errExpected:          true,
			errSubstringExpected: ErrCannotOpenOrCreateDstFile.Error(),
		},
		{
			name:                 "atomic_dst_path_which_cannot_be_opened_or_created",
			inFile:               tmpFile(),
			outFile:              "/path/that/is/inaccessible",
			errExpected:          true,
			options:              Options{Atomic: true},
			errSubstringExpected: ErrCannotCreateTmpFile.Error(),
		},
		{
			name:    "src_directory_when_recurse_is_set_false",
			inFile:  tmpDirPath(),
			outFile: tmpDirPath(),
			options: Options{
				Recursive: false,
			},
			errExpected:          true,
			errSubstringExpected: ErrOmittingDir.Error(),
		},
		{
			name:                 "verify_cannot_overwrite_file_with_dir",
			inFile:               tmpDirPath(),
			outFile:              tmpFile(),
			options:              Options{Recursive: true},
			errExpected:          true,
			errSubstringExpected: ErrCannotOverwriteNonDir.Error(),
		},
		{
			name:                 "err_when_parents_true_but_dst_is_not_dir",
			inFile:               tmpDirPath(),
			outFile:              tmpFile(),
			options:              Options{Parents: true},
			errExpected:          true,
			errSubstringExpected: ErrWithParentsDstMustBeDir.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Copy(tt.inFile, tt.outFile, tt.options)
			if tt.errExpected {
				assert.True(errContains(err, tt.errSubstringExpected), fmt.Sprintf("err '%s' does not contain '%s'", err, tt.errSubstringExpected))
			} else {
				assert.Nil(err)
			}
		})
	}
}

func TestCopyingBasicFileContentsWhenDstFileExists(t *testing.T) {
	assert := assert.New(t)
	src, dst := tmpFile(), tmpFile()
	content := []byte("foo")

	err := ioutil.WriteFile(src, content, 0655)
	assert.Nil(err)

	err = SimpleCopy(src, dst)
	assert.Nil(err)

	b, err := ioutil.ReadFile(dst)
	assert.Nil(err)
	assert.Equal(content, b)
}

func TestCopyingBasicFileContentsWhenDstFileDoesNotExist(t *testing.T) {
	assert := assert.New(t)
	src := tmpFile()
	dst := tmpFilePathUnused()
	content := []byte("foo")
	err := ioutil.WriteFile(src, content, 0655)
	assert.Nil(err)

	err = SimpleCopy(src, dst)
	assert.Nil(err)

	b, err := ioutil.ReadFile(dst)
	assert.Nil(err)
	assert.Equal(content, b)
}

func TestCopyingSameFileReturnsNoError(t *testing.T) {
	assert := assert.New(t)
	tmp := tmpFile()
	err := SimpleCopy(tmp, tmp)
	assert.Nil(err)
}

func TestCopySrcFileToDstDir(t *testing.T) {
	assert := assert.New(t)
	src := tmpFile()
	content := []byte("foo")

	tests := []struct {
		name                 string
		dst                  string
		options              Options
		errExpected          bool
		errSubstringExpected string
	}{
		{name: "auto_write_files_to_dirs", dst: tmpDirPath(), options: Options{AppendNameToPath: true}},
		{name: "auto_write_files_to_dirs", dst: tmpDirPath(), options: Options{AppendNameToPath: false}, errExpected: true, errSubstringExpected: ErrWritingFileToExistingDir.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ioutil.WriteFile(src, content, 0655)
			assert.Nil(err)

			// make sure we don't error
			err = Copy(src, tt.dst, tt.options)
			if tt.errExpected {
				assert.True(errContains(err, tt.errSubstringExpected))
			} else {
				assert.Nil(err)
				// make sure the dst inFile is in the dir where we expect it
				b, err := ioutil.ReadFile(filepath.Join(tt.dst, filepath.Base(src)))
				assert.Nil(err)
				assert.Equal(content, b)
			}
		})
	}
}

func TestCopyingDirToDirWhenDstDirDoesNotExist(t *testing.T) {
	assert := assert.New(t)
	src, dst := tmpDirPath(), tmpDirPathUnused()

	// make files with content
	file1 := filepath.Join(src, "file1.txt")
	assert.Nil(os.Mkdir(filepath.Join(src, "subdir"), 0777))
	file3 := filepath.Join(src, "subdir", "file2.txt")
	files := []string{file1, file3}

	// write content to files
	content := []byte("foo")
	for _, file := range files {
		assert.Nil(ioutil.WriteFile(file, content, 0655))
	}

	// ensure src exists
	_, err := os.Open(src)
	assert.False(os.IsNotExist(err))

	// copy directory to directory recursively
	assert.Nil(Copy(src, dst, Options{Recursive: true}))

	// verify the file contents
	assert.Nil(err)
	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		assert.Nil(err)
		assert.Equal(content, b)
	}
}

func TestCopyingDirToDirWhenDstContainsTrailingSlash(t *testing.T) {
	// at some point we try to determine if the dst directory's parent
	// exists. when given /some/path the parent is /some.  when given
	// /some/path/ the parent is /some/path.  this means the destination
	// dir is not created when it should be.  this tests for that case.

	assert := assert.New(t)
	src, dst := tmpDirPath(), tmpDirPathUnused()+"/"

	// make files with content
	file1 := filepath.Join(src, "file1.txt")
	assert.Nil(os.Mkdir(filepath.Join(src, "subdir"), 0777))
	file3 := filepath.Join(src, "subdir", "file2.txt")
	files := []string{file1, file3}

	// write content to files
	content := []byte("foo")
	for _, file := range files {
		assert.Nil(ioutil.WriteFile(file, content, 0655))
	}

	// ensure src exists
	_, err := os.Open(src)
	assert.False(os.IsNotExist(err))

	// copy directory to directory recursively
	assert.Nil(Copy(src, dst, Options{Recursive: true}))

	// verify the file contents
	assert.Nil(err)
	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		assert.Nil(err)
		assert.Equal(content, b)
	}
}

func TestCopyingFileWithParentsFlag(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name    string
		options Options
		// should the src be a directory?
		srcIsDir bool
		// should the dst be a directory?
		dstIsDir             bool
		errSubstringExpected string
	}{
		{name: "src_file_dst_file_with_parents_option_errors", srcIsDir: false, dstIsDir: false, errSubstringExpected: ErrWithParentsDstMustBeDir.Error()},
		{name: "src_dir_dst_file_with_parents_option_errors", srcIsDir: true, dstIsDir: false, errSubstringExpected: ErrWithParentsDstMustBeDir.Error()},
		{name: "src_file_dst_dir", srcIsDir: false, dstIsDir: true},
		{name: "src_file_dst_dir_with_append", srcIsDir: false, dstIsDir: true, options: Options{AppendNameToPath: true}, errSubstringExpected: ErrIncompatibleOptions.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := tmpDirPath()
			nestedPath := "nested/path/" // nested path in the local working dir
			{
				defer os.RemoveAll(workDir)
				// move into working directory to start with a clean slate
				err := os.Chdir(workDir)
				assert.Nil(err)

				assert.Nil(os.MkdirAll(nestedPath, 0777))
				defer os.RemoveAll(nestedPath)
			}

			// make files with content in nested path
			content := []byte("foo")
			fileName := "src.txt"
			var src string
			if tt.srcIsDir {
				src = nestedPath
				// write to file inside dir while maintaining nested dir as src
				assert.Nil(ioutil.WriteFile(filepath.Join(src, fileName), content, 0655))

				// ensure nested exists just to be sure
				f, err := os.Open(filepath.Join(src, fileName))
				assert.Nil(err)
				// ensure file exists
				assert.False(os.IsNotExist(err))
				assert.Nil(f.Close())
			} else {
				src = filepath.Join(nestedPath, fileName)

				// write content to file which we will verify later
				assert.Nil(ioutil.WriteFile(src, content, 0655))

				// ensure nested exists just to be sure
				f, err := os.Open(src)
				assert.Nil(err)
				// ensure file exists
				assert.False(os.IsNotExist(err))
				assert.Nil(f.Close())
			}

			var dst string
			if tt.dstIsDir {
				// make destination dir
				dst = "dst"
				assert.Nil(os.Mkdir(dst, 0777))
				defer os.RemoveAll(dst)

			} else {
				// make destination file
				dst = filepath.Join(tmpDirPath(), "file.txt")
			}

			// add required options to those given
			{
				tt.options.Parents = true
				tt.options.Recursive = true
				tt.options.DebugLogFunc = debugLogger
				tt.options.InfoLogFunc = infoLogger
			}

			a, _ := filepath.Abs(dst)
			fmt.Println("ABS:", a) // FIXME: testing
			err := Copy(src, dst, tt.options)

			// check the err from copy
			if len(tt.errSubstringExpected) > 0 {
				assert.True(errContains(err, tt.errSubstringExpected), fmt.Sprintf("err '%s' does not contain '%s'", err, tt.errSubstringExpected))
				return
			} else {
				assert.Nil(err)
			}

			expectedFile := filepath.Join(dst, nestedPath, fileName)

			// ensure the file exists where we expect it to
			var exists bool
			if _, err = os.Stat(expectedFile); !os.IsNotExist(err) {
				exists = true
			}
			assert.True(exists)
		})
	}
}

func TestNoClobberFile(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name string
		opts Options
		// do we expect the dst file to be overwritten?
		expectOverwrite bool
	}{
		{name: "expect_clobber", opts: Options{NoClobber: false}, expectOverwrite: true},
		{name: "basic_no_clobber", opts: Options{NoClobber: true}, expectOverwrite: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcContent := []byte("source")
			dstContent := []byte("dest")

			src, dst := tmpFile(), tmpFile()

			assert.Nil(ioutil.WriteFile(src, srcContent, 0655))
			assert.Nil(ioutil.WriteFile(dst, dstContent, 0655))

			assert.Nil(Copy(src, dst, tt.opts))
			b, err := ioutil.ReadFile(dst)
			assert.Nil(err)
			if tt.expectOverwrite {
				assert.Equal(srcContent, b)
			} else {
				assert.Equal(dstContent, b)
			}
		})
	}
}

func TestNoErrWhenCopyfileAlreadyExists(t *testing.T) {
	assert := assert.New(t)
	src, dst := tmpFile(), tmpFile()
	dstDir := filepath.Dir(dst)

	// use the known expected name for a copy file
	copyFileName := "copyfile-"
	copyFileFullyQualified := filepath.Join(dstDir, copyFileName)
	assert.Nil(ioutil.WriteFile(copyFileFullyQualified, []byte("foo"), 0655))
	// hold the copy file open to give our atomic copy code problems
	f, err := os.Open(copyFileFullyQualified)
	assert.Nil(err)

	assert.Nil(Copy(src, dst, Options{Atomic: true}))
	_ = f.Close()
}

func TestCreatingSimpleBackupFile(t *testing.T) {
	assert := assert.New(t)
	src, dst := tmpFile(), tmpFile()

	content := []byte("foo")
	assert.Nil(ioutil.WriteFile(dst, content, 0655))

	assert.Nil(Copy(src, dst, Options{Backup: "simple"}))
	expectedBkpFile := dst + "~"

	_, err := os.Stat(expectedBkpFile)
	assert.Nil(err)
	assert.False(os.IsNotExist(err))

	b, err := ioutil.ReadFile(expectedBkpFile)
	assert.Nil(err)
	assert.Equal(content, b)
}

func TestCreatingExistingBackupFilesWhenBackupIsNotPresent(t *testing.T) {
	assert := assert.New(t)
	src, dst := tmpFile(), tmpFile()

	content := []byte("foo")
	assert.Nil(ioutil.WriteFile(dst, content, 0655))

	assert.Nil(Copy(src, dst, Options{Backup: "existing"}))

	expectedBkpFile := dst + "~"
	_, err := os.Stat(expectedBkpFile)
	assert.Nil(err)
	assert.False(os.IsNotExist(err))

	b, err := ioutil.ReadFile(expectedBkpFile)
	assert.Nil(err)
	assert.Equal(content, b)
}

func TestNumberedBackupFileRegex(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		in          string
		expectMatch bool
		msg         string
	}{
		{"f.txt.~1~", true, ""},
		{"f.txt~1~", false, "missing . before the ~"},
		{"f.txt.~12~", true, ""},
		{"f.txt.~12345~", true, ""},
		{"f.txt.~123456789~", false, "we don't expect to see numbers this long"},
		{"f.txt.~1x5~", false, "alpha in between numbers"},
		{"f.txt~", false, ""},
		{"f.t123xt", false, ""},
		{"123", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(tt.expectMatch, numberedBackupFile.MatchString(tt.in), tt.msg)
		})
	}
}

func TestNumberedBackupRegexSubstring(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		in  string
		sub string // we expect to see this substring in index 1
	}{
		{"f.txt.~1~", "1"},
		{"f.txt.~12~", "12"},
		{"123.456.~78~", "78"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			sub := numberedBackupFile.FindStringSubmatch(tt.in)
			if len(sub) == 0 {
				assert.FailNow("substring has length 0")
			}
			if len(sub) <= 1 {
				assert.FailNow("substring was not greater than 1 as we expect")
			}
			assert.Equal(string(sub[1]), tt.sub)
		})
	}
}

//noinspection ALL
func TestCreatingNumberedBackupFilesWhenBackupFilesAreNumbered(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name                  string
		numberedFilesToCreate []int
		expectedFileNum       int
	}{
		{"single_backup_file", []int{1}, 2},
		{"multiple_files", []int{1, 2}, 3},
		{"skip_files", []int{1, 3}, 4},
		{"skip_first_num", []int{2, 3}, 4},
		{"larger_numbers_just_for_good_measure", []int{5, 7, 5431}, 5432},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tmpFile(), tmpFile()
			content := []byte("foo")
			assert.Nil(ioutil.WriteFile(dst, content, 0655))
			defer os.Remove(dst)

			// write all of the numbered files so we have a starting point
			for _, num := range tt.numberedFilesToCreate {
				numberedFile := fmt.Sprintf("%s.~%d~", dst, num)
				assert.Nil(ioutil.WriteFile(numberedFile, content, 0655))
				defer os.Remove(numberedFile)
			}

			assert.Nil(Copy(src, dst, Options{Backup: "numbered"}))
			expectedBkpFile := fmt.Sprintf("%s.~%d~", dst, tt.expectedFileNum)
			b, err := ioutil.ReadFile(expectedBkpFile)
			assert.Nil(err)
			assert.Equal(content, b)
		})
	}
}

// ensure that we are properly cleaning file paths before use
func TestCleaningFilePaths(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"src_trailing_forward_slashes", tmpFile() + "/////", tmpFile()},
		{"dst_trailing_forward_slashes", tmpFile(), tmpFile() + "////////"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(SimpleCopy(tt.in, tt.out))
		})
	}
}
