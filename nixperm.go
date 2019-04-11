// +build linux darwin

package flop

import (
	"os"
	"syscall"
	"time"
)

// setAttributes will set file level permissions and attributes on dst based on Preserve options.
func setAttributes(srcFile File, dstFile File, opts Options) error {
	// set mode
	if opts.Preserve.Mode {
		mode := srcFile.Mode()
		opts.logDebug("setting dst %s to src permissions %s", dstFile.Path, mode)
		if err := os.Chmod(dstFile.Path, mode); err != nil {
			return err
		}
	}

	// set timestamps
	if opts.Preserve.Timestamps {
		// get src times
		fileInfo, err := os.Stat(srcFile.Path)
		if err != nil {
			opts.logInfo(ErrCannotStatFile.Error(), "src file %s", srcFile.Path)
			return err
		}
		statT := fileInfo.Sys().(*syscall.Stat_t)
		srcATime := time.Unix(statT.Atim.Sec, statT.Atim.Nsec)
		srcMTime := time.Unix(statT.Mtim.Sec, statT.Mtim.Nsec)

		// set dst times
		os.Chtimes(dstFile.Path, srcATime, srcMTime)
	}

	if opts.Preserve.ownership {
		// get src ownership
		fileInfo, err := os.Stat(srcFile.Path)
		if err != nil {
			opts.logInfo(ErrCannotStatFile.Error(), "src file %s", srcFile.Path)
			return err
		}
		statT := fileInfo.Sys().(*syscall.Stat_t)

		// set dst ownership
		if err := os.Chown(dstFile.Path, int(statT.Uid), int(statT.Gid)); err != nil {
			opts.logInfo(ErrCannotChownFile.Error(), "dst file %s, if this is a permission error consider setting PreserveAttr None to avoid setting Ownership", dstFile.Path)
			return err
		}
	}
	return nil
}
