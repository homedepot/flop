package flop

import (
	"fmt"
)

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
	// is returned is destination is not a directory.
	Parents bool
	// Preserve attributes like mode and timestamps.  Reference PreserveAttrs for details.
	Preserve PreserveAttrs
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

// Preserve the specified attributes. Accepted values are:
//   - "mode"        preserve the file mode bits and access control lists
//   - "ownership"   preserve the owner and group if permissions allow
//   - "timestamps"  preserve times of last access and last modification, when possible
//   - future support: "context"  preserve SELinux contexts
//   - future support: "links"    preserve in the destination files any links between corresponding source files
//   - future support: "xattr"    preserve extended attributes of the file
//   - "all"         preserve all supported attributes
//
//
// Not setting Preserve at all is equivalent to []string{"mode"}

// PreserveAttrs stores Options.Preserve attributes.  Using Preserve with no attributes is equivalent to setting
// Mode, Ownership and Timestamps.
type PreserveAttrs struct {
	// All marks all of the attributes here for preservation.
	All bool
	// None will explicitly ignore preserving any attributes.
	None bool
	// Mode preserves the file mode bits.
	Mode bool // TODO: implement the 'and access control lists' part of this attribute, and add back to description.
	// ownership preserves the owner and group if permissions allow.
	ownership bool // future support
	// Timestamps preserves times of last access and last modification, when possible.
	Timestamps bool
	// context preserves SELinux contexts.
	context bool // future support
	// links preserves in the destination files any links between corresponding source files.
	links bool // future support
	// xattr preserves extended attributes of the file.
	xattr bool // future support
}

// undefined returns true when none of the attributes have been set.
func (a *PreserveAttrs) undefined() bool {
	if a.All == false &&
		a.None == false &&
		a.Mode == false &&
		a.ownership == false &&
		a.Timestamps == false &&
		a.context == false &&
		a.links == false &&
		a.xattr == false {
		return true
	}
	return false
}

// setDefaults marks the default preserve attributes as true.
func (a *PreserveAttrs) setDefaults() {
	a.Mode = true
	a.ownership = true
	a.Timestamps = true
}

// setAll marks all preserve attributes as true, save None.
func (a *PreserveAttrs) setAll() {
	a.All = true
	a.Mode = true
	a.ownership = true
	a.Timestamps = true
	a.context = true
	a.links = true
	a.xattr = true
}
