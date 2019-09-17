package flop

import "fmt"

// Options directly represent command line flags associated with GNU file operations.
type Options struct {
	// AppendNameToPath will, when attempting to copy a file to an existing directory, automatically
	// create the file with the same name in the destination directory.  While CP uses this behavior
	// by default it is an assumption better left to the client in a programmatic setting.
	AppendNameToPath bool
	// Atomic will copy contents to a temporary file in the destination's parent directory first, then
	// rename the file to ensure the operation is atomic.
	Atomic bool
	// Backup makes a backup of each existing destination file. The backup suffix is '~'. Acceptable
	// control values are:
	//   - "off"       no backup will be made (default)
	//   - "simple"    always make simple backups
	//   - "numbered"  make numbered backups
	//   - "existing"  numbered if numbered backups exist, simple otherwise
	Backup string
	// Link creates hard links to files instead of copying them.
	Link bool
	// MkdirAll will use os.MkdirAll to create the destination directory if it does not exist, along with
	// any necessary parents.
	MkdirAll bool
	// mkdirAll is an internal tracker for MkdirAll, including other validation checks
	mkdirAll bool
	// NoClobber will not let an existing file be overwritten.
	NoClobber bool
	// Parents will create source directories in dst if they do not already exist. ErrWithParentsDstMustBeDir
	// is returned if destination is not a directory.
	Parents bool
	// Recursive will recurse through sub directories if set true.
	Recursive bool
	// InfoLogFunc will, if defined, handle logging info messages.
	InfoLogFunc func(string)
	// DebugLogFunc will, if defined, handle logging debug messages.
	DebugLogFunc func(string)
}

// setLoggers will configure logging functions, setting noop loggers if log funcs are undefined.
func (o *Options) setLoggers() {
	if o.InfoLogFunc == nil {
		o.InfoLogFunc = func(string) {}
	}
	if o.DebugLogFunc == nil {
		o.DebugLogFunc = func(string) {}
	}
}

// logDebug will log to the DebugLogFunc.
func (o *Options) logDebug(format string, a ...interface{}) {
	o.DebugLogFunc(fmt.Sprintf(format, a...))
}

// logInfo will log to the InfoLogFunc.
func (o *Options) logInfo(format string, a ...interface{}) {
	o.InfoLogFunc(fmt.Sprintf(format, a...))
}
