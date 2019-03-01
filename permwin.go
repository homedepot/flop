// +build windows

package flop

import (
	"os"
)

// setPermissions on Windows systems is a noop.  This will need to be handled by the client.
func setPermissions(dstFile *File, srcMode os.FileMode, opts Options) error {
	opts.logDebug("permission handling is ignored on Windows, dst file %s will be unchanged", dstFile.Path)
	return nil
}
