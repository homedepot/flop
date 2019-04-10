// +build linux darwin

package flop

import (
	"github.com/pkg/errors"
	"os"
)

// ensurePermissions will set file level permissions on dst based on options and other criteria.
func ensurePermissions(dstFile File, srcMode os.FileMode, opts Options) error {
	var mode os.FileMode
	if fi, err := os.Stat(dstFile.Path); err != nil {
		return err
	} else {
		mode = fi.Mode()
	}

	if dstFile.Exists() {
		if mode == dstFile.Mode() {
			opts.logDebug("existing dst %s permissions %s are unchanged", dstFile.Path, mode)
			return nil
		}

		// make sure dst perms are set to their original value
		opts.logDebug("changing dst %s permissions to %s", dstFile.Path, dstFile.Mode())
		err := os.Chmod(dstFile.Path, dstFile.Mode())
		if err != nil {
			return errors.Wrapf(ErrCannotChmodFile, "destination file %s: %s", dstFile.Path, err)
		}
	} else {
		if mode == srcMode {
			opts.logDebug("dst %s permissions %s already match src perms", dstFile.Path, mode)
		}

		// make sure dst perms are set to that of src
		opts.logDebug("changing dst %s permissions to %s", dstFile.Path, srcMode)
		err := os.Chmod(dstFile.Path, srcMode)
		if err != nil {
			return errors.Wrapf(ErrCannotChmodFile, "destination file %s: %s", dstFile.Path, err)
		}
	}
	return nil
}
