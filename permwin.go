// +build windows

package flop

import (
	"os"
)

// ensurePermissions on Windows systems is a noop.  This will need to be handled by the client.
func ensurePermissions(dstFile *File, srcMode os.FileMode, opts Options) error {
	opts.logDebug("permission handling is ignored on Windows, dst file %s will be unchanged", dstFile.Path)
	return nil
}
