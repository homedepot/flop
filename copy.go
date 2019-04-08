// Package flop implements file operations, taking most queues from GNU cp while trying to be
// more programmatically friendly.
package flop

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// numberedBackupFile matches files that looks like file.ext.~1~ and uses a capture group to grab the number
var numberedBackupFile = regexp.MustCompile(`^.*\.~([0-9]{1,5})~$`)

// File describes a file and associated options for operations on the file.
type File struct {
	// Path is the path to the src file.
	Path string
	// fileInfoOnInit is os.FileInfo for file when initialized.
	fileInfoOnInit os.FileInfo
	// existOnInit is true if the file exists when initialized.
	existOnInit bool
	// isDir is true if the file object is a directory.
	isDir bool
}

// NewFile creates a new File.
func NewFile(path string) *File {
	return &File{
		Path: path,
	}
}

// setInfo will collect information about a File and populate the necessary fields.
func (f *File) setInfo() error {
	info, err := os.Lstat(f.Path)
	f.fileInfoOnInit = info
	if err != nil {
		if !os.IsNotExist(err) {
			// if we are here then we have an error, but not one indicating the file does not exist
			return err
		}
	} else {
		f.existOnInit = true
		if f.fileInfoOnInit.IsDir() {
			f.isDir = true
		}
	}
	return nil
}

func (f *File) isSymlink() bool {
	if f.fileInfoOnInit.Mode()&os.ModeSymlink != 0 {
		return true
	}
	return false
}

// shouldMakeParents returns true if we should make parent directories up to the dst
func (f *File) shouldMakeParents(opts Options) bool {
	if opts.MkdirAll || opts.mkdirAll {
		return true
	}

	if opts.Parents {
		return true
	}

	if f.existOnInit {
		return false
	}

	parent := filepath.Dir(filepath.Clean(f.Path))
	if _, err := os.Stat(parent); !os.IsNotExist(err) {
		// dst does not exist but the direct parent does. make the target dir.
		return true
	}

	return false
}

// shouldCopyParents returns true if parent directories from src should be copied into dst.
func (f *File) shouldCopyParents(opts Options) bool {
	if !opts.Parents {
		return false
	}
	return true
}

// SimpleCopy will src to dst with default Options.
func SimpleCopy(src, dst string) error {
	return Copy(src, dst, Options{})
}

// Copy will copy src to dst.  Behavior is determined by the given Options.
func Copy(src, dst string, opts Options) (err error) {
	opts.setLoggers()
	srcFile, dstFile := NewFile(filepath.Clean(src)), NewFile(filepath.Clean(dst))

	// set src attributes
	if err := srcFile.setInfo(); err != nil {
		return errors.Wrapf(ErrCannotStatFile, "source file %s: %s", srcFile.Path, err)
	}
	if !srcFile.existOnInit {
		return errors.Wrapf(ErrFileNotExist, "source file %s", srcFile.Path)
	}
	opts.logDebug("src %s existOnInit: %t", srcFile.Path, srcFile.existOnInit)

	// stat dst attributes. handle errors later
	_ = dstFile.setInfo()
	opts.logDebug("dst %s existOnInit: %t", dstFile.Path, dstFile.existOnInit)

	if dstFile.shouldMakeParents(opts) {
		opts.mkdirAll = true
		opts.DebugLogFunc("dst mkdirAll: true")
	}

	if opts.Parents {
		if !dstFile.IsDir() {
			return ErrWithParentsDstMustBeDir
		}
		// TODO: figure out how to handle windows paths where they reference the full path like c:/dir
		dstFile.Path = filepath.Join(dstFile.Path, srcFile.Path)
		opts.logDebug("because of Parents option, setting dst Path to %s", dstFile.Path)
		dstFile.isDir = srcFile.isDir
		opts.Parents = false // ensure we don't keep creating parents on recursive calls
	}

	// copying src directory requires dst is also a directory, if it existOnInit
	if srcFile.isDir && dstFile.existOnInit && !dstFile.isDir {
		return errors.Wrapf(
			ErrCannotOverwriteNonDir, "source directory %s, destination file %s", srcFile.Path, dstFile.Path)
	}

	// divide and conquer
	switch {
	case opts.Link:
		return hardLink(srcFile, dstFile, opts.logDebug)
	case srcFile.isSymlink():
		// FIXME: we really need to copy the pass through dest unless they specify otherwise...check the docs
		return copyLink(srcFile, dstFile, opts.logDebug)
	case srcFile.isDir:
		return copyDir(srcFile, dstFile, opts)
	default:
		return copyFile(srcFile, dstFile, opts)
	}
}

// hardLink creates a hard link to src at dst.
func hardLink(src, dst *File, logFunc func(format string, a ...interface{})) error {
	logFunc("creating hard link to src %s at dst %s", src.Path, dst.Path)
	return os.Link(src.Path, dst.Path)
}

// copyLink copies a symbolic link from src to dst.
func copyLink(src, dst *File, logFunc func(format string, a ...interface{})) error {
	logFunc("copying sym link %s to %s", src.Path, dst.Path)
	linkSrc, err := os.Readlink(src.Path)
	if err != nil {
		return err
	}
	return os.Symlink(linkSrc, dst.Path)
}

func copyDir(srcFile, dstFile *File, opts Options) error {
	if !opts.Recursive {
		return errors.Wrapf(ErrOmittingDir, "source directory %s", srcFile.Path)
	}
	if opts.mkdirAll {
		opts.logDebug("making all dirs up to %s", dstFile.Path)
		if err := os.MkdirAll(dstFile.Path, srcFile.fileInfoOnInit.Mode()); err != nil {
			return err
		}
	}

	srcDirEntries, err := ioutil.ReadDir(srcFile.Path)
	if err != nil {
		return errors.Wrapf(ErrReadingSrcDir, "source directory %s: %s", srcFile.Path, err)
	}

	for _, entry := range srcDirEntries {
		newSrc := filepath.Join(srcFile.Path, entry.Name())
		newDst := filepath.Join(dstFile.Path, entry.Name())
		opts.logDebug("recursive cp with src %s and dst %s", newSrc, newDst)
		if err := Copy(
			newSrc,
			newDst,
			opts,
		); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(srcFile, dstFile *File, opts Options) (err error) {
	// shortcut if files are the same file
	if os.SameFile(srcFile.fileInfoOnInit, dstFile.fileInfoOnInit) {
		opts.logDebug("src %s is same file as dst %s", srcFile.Path, dstFile.Path)
		return nil
	}

	// optionally make dst parent directories
	if dstFile.shouldMakeParents(opts) {
		// TODO: permissive perms here to ensure tmp file can write on nix.. ensure we are setting these correctly down the line or fix here
		if err := os.MkdirAll(filepath.Dir(dstFile.Path), 0777); err != nil {
			return err
		}
	}

	if dstFile.existOnInit {
		if dstFile.isDir {
			// optionally append src file name to dst dir like cp does
			if opts.AppendNameToPath {
				dstFile.Path = filepath.Join(dstFile.Path, filepath.Base(srcFile.Path))
				opts.logDebug("because of AppendNameToPath option, setting dst path to %s", dstFile.Path)
			} else {
				return errors.Wrapf(ErrWritingFileToExistingDir, "destination directory %s", dstFile.Path)
			}
		}

		// optionally do not clobber existing dst file
		if opts.NoClobber {
			opts.logDebug("dst %s exists, will not clobber", dstFile.Path)
			return nil
		}

		if opts.Backup != "" {
			if err := backupFile(dstFile, opts.Backup, opts); err != nil {
				return err
			}
		}

	}

	srcFD, err := os.Open(srcFile.Path)
	if err != nil {
		return errors.Wrapf(ErrCannotOpenSrc, "source file %s: %s", srcFile.Path, err)
	}
	defer func() {
		if closeErr := srcFD.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	if opts.Atomic {
		dstDir := filepath.Dir(dstFile.Path)
		tmpFD, err := ioutil.TempFile(dstDir, "copyfile-")
		defer closeAndRemove(tmpFD, opts.logDebug)
		if err != nil {
			return errors.Wrapf(ErrCannotCreateTmpFile, "destination directory %s: %s", dstDir, err)
		}
		opts.logDebug("created tmp file %s", tmpFD.Name())

		//copy src to tmp and cleanup on any error
		opts.logInfo("copying src file %s to tmp file %s", srcFD.Name(), tmpFD.Name())
		if _, err := io.Copy(tmpFD, srcFD); err != nil {
			return err
		}
		if err := tmpFD.Sync(); err != nil {
			return err
		}
		if err := tmpFD.Close(); err != nil {
			return err
		}

		// move tmp to dst
		opts.logInfo("renaming tmp file %s to dst %s", tmpFD.Name(), dstFile.Path)
		if err := os.Rename(tmpFD.Name(), dstFile.Path); err != nil {
			return errors.Wrapf(ErrCannotRenameTempFile, "attempted to rename temp transfer file %s to %s", tmpFD.Name(), dstFile.Path)
		}
	} else {
		dstFD, err := os.Create(dstFile.Path)
		if err != nil {
			return errors.Wrapf(ErrCannotOpenOrCreateDstFile, "destination file %s: %s", dstFile.Path, err)
		}

		opts.logInfo("copying src file %s to dst file %s", srcFD.Name(), dstFD.Name())
		if _, err = io.Copy(dstFD, srcFD); err != nil {
			return err
		}
		if err := dstFD.Sync(); err != nil {
			return err
		}
	}

	return setPermissions(dstFile, srcFile.fileInfoOnInit.Mode(), opts)
}

// backupFile will create a backup of the file using the chosen control method.  See Options.Backup.
func backupFile(file *File, control string, opts Options) error {
	// TODO: this func could be more efficient if it used file instead of the path but right now this causes panic
	// do not copy if the file did not exist
	if !file.existOnInit {
		return nil
	}

	// simple backup
	simple := func() error {
		bkp := file.Path + "~"
		opts.logDebug("creating simple backup file %s", bkp)
		return Copy(file.Path, bkp, opts)
	}

	// next gives the next unused backup file number, 1 above the current highest
	next := func() (int, error) {
		// find general matches that look like numbered backup files
		m, err := filepath.Glob(file.Path + ".~[0-9]*~")
		if err != nil {
			return -1, err
		}

		// get each backup file num substring, convert to int, track highest num
		var highest int
		for _, f := range m {
			subs := numberedBackupFile.FindStringSubmatch(filepath.Base(f))
			if len(subs) > 1 {
				if i, _ := strconv.Atoi(string(subs[1])); i > highest {
					highest = i
				}
			}
		}
		return highest + 1, nil
	}

	// numbered backup
	numbered := func(n int) error {
		return Copy(file.Path, fmt.Sprintf("%s.~%d~", file.Path, n), opts)
	}

	switch control {
	default:
		return errors.Wrapf(ErrInvalidBackupControlValue, "backup value '%s'", control)
	case "off":
		return nil
	case "simple":
		return simple()
	case "numbered":
		i, err := next()
		if err != nil {
			return err
		}
		return numbered(i)
	case "existing":
		i, err := next()
		if err != nil {
			return err
		}

		if i > 1 {
			return numbered(i)
		}
		return simple()
	}
}

func closeAndRemove(file *os.File, logFunc func(format string, a ...interface{})) {
	if file != nil {
		if err := file.Close(); err != nil {
			logFunc("err closing file %s: %s", file.Name(), err)
		}
		if err := os.Remove(file.Name()); err != nil {
			logFunc("err removing file %s: %s", file.Name(), err)
		}
	}
}
