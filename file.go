package flop

import (
	"os"
	"path/filepath"
)

// File describes a file and associated options for operations on the file.
type File struct {
	// Path is the path to the src file.
	Path string
	// fileInfo is an internal tracker for fileInfo to eliminate statting a file more than once.
	fileInfo *os.FileInfo
	// exists is an internal tracker for file existence. use Exist() to get file existence.
	exists *bool
	// isDir is an internal tracker for whether a file is a directory.  use IsDir() to find
	// directory status
	isDir *bool
}

// NewFile creates a new File.
func NewFile(path string) *File {
	return &File{Path: path}
}

// Exists returns true if we can guarantee file existence.  Other errors, like invalid parent directory permissions,
// are swallowed and false returned.
func (f *File) Exists() bool {
	if f.exists != nil {
		return *f.exists
	}

	var b bool
	if _, err := os.Lstat(f.Path); err == nil {
		b = true
	} else {
		b = false
	}
	f.exists = &b
	return *f.exists
}

// stat will os.Lstat a file and save os.FileInfo.
func (f *File) stat() error {
	fi, err := os.Lstat(f.Path)
	if err != nil {
		return err
	}
	f.fileInfo = &fi
	return nil
}

// FileInfo returns os.FileInfo for the file, using stored fileInfo to avoid re-statting.
func (f *File) FileInfo() os.FileInfo {
	if f.fileInfo == nil {
		_ = f.stat()
	}

	if f.fileInfo == nil {
		var fi os.FileInfo
		return fi
	}
	return *f.fileInfo
}

// IsDir returns true if the File is a directory.
func (f *File) IsDir() bool {
	if f.isDir != nil {
		return *f.isDir
	}

	if f.fileInfo == nil {
		f.stat()
	}

	var b bool
	if f.fileInfo != nil && (*f.fileInfo).IsDir() {
		b = true
	} else {
		b = false
	}
	f.isDir = &b
	return *f.isDir
}

// Mode returns file mode bits.
func (f *File) Mode() os.FileMode {
	if f.fileInfo == nil {
		f.stat() // FIXME: handle this error appropriately
	}
	// make sure we don't try to call Mode() if fileInfo is still nil.
	if f.fileInfo == nil {
		var m os.FileMode
		return m
	}
	return (*f.fileInfo).Mode()
}

func (f *File) isSymlink() bool {
	if f.Mode()&os.ModeSymlink != 0 {
		return true
	}
	return false
}

// TODO: move this elsewhere.. too much logic from other stuffs
// shouldMakeParents returns true if we should make parent directories up to the dst
func (f *File) shouldMakeParents(opts Options) bool {
	if opts.MkdirAll || opts.mkdirAll {
		return true
	}

	if opts.Parents {
		return true
	}

	if f.Exists() {
		return false
	}

	//parent := filepath.Dir(filepath.Clean(f.Path))  // FIXME: which of these is right.
	parent := filepath.Dir(f.Path) // FIXME: <- this one from changes
	if _, err := os.Stat(parent); !os.IsNotExist(err) {
		// dst does not exist but the direct parent does. make the target dir.
		return true
	}

	return false
}

// TODO: move this elsewhere.. too much logic from other stuffs
// shouldCopyParents returns true if parent directories from src should be copied into dst.
func (f *File) shouldCopyParents(opts Options) bool {
	if !opts.Parents {
		return false
	}
	return true
}
