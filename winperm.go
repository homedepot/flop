// +build windows

package flop

// setAttributes on Windows systems is a noop.  This will need to be handled by the client.
func setAttributes(srcFile File, dstFile File, opts Options) error {
	opts.logDebug("permission handling is ignored on Windows, dst file %s will be unchanged", dstFile.Path)
	return nil
}
