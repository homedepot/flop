package flop

import "github.com/pkg/errors"

var (
	// ErrFileNotExist occurs when a file is given that does not exist when its existence is required.
	ErrFileNotExist = errors.New("no such file or directory")
	// ErrCannotOpenSrc occurs when a src file cannot be opened with os.Open().
	ErrCannotOpenSrc = errors.New("source file cannot be opened")
	// ErrCannotStatFile occurs when a file receives an error from get os.Stat().
	ErrCannotStatFile = errors.New("cannot stat file, check that file path is accessible") // TODO: identify a location for this error or remove
	// ErrCannotChmodFile occurs when an error is received trying to change permissions on a file.
	ErrCannotChmodFile = errors.New("cannot change permissions on file")
	// ErrCannotChownFile occurs when an error is received trying to change ownership on a file.
	ErrCannotChownFile = errors.New("cannot change ownership on file")
	// ErrCannotCreateTmpFile occurs when an error is received attempting to create a temporary file for atomic copy.
	ErrCannotCreateTmpFile = errors.New("temp file cannot be created")
	// ErrCannotOpenOrCreateDstFile occurs when an error is received attempting to open or create destination file during non-atomic copy.
	ErrCannotOpenOrCreateDstFile = errors.New("destination file cannot be created")
	// ErrCannotRenameTempFile occurs when an error is received trying to rename the temporary copy file to the destination.
	ErrCannotRenameTempFile = errors.New("cannot rename temp file, check file or directory permissions")
	// ErrOmittingDir occurs when attempting to copy a directory but Options.Recursive is not set to true.
	ErrOmittingDir = errors.New("Options.Recursive is not true, omitting directory")
	// ErrWithParentsDstMustBeDir occurs when the destination is expected to be an existing directory but is not
	// present or accessible.
	ErrWithParentsDstMustBeDir = errors.New("with Options.Parents, the destination must be a directory")
	// ErrCannotOverwriteNonDir occurs when attempting to copy a directory to a non-directory.
	ErrCannotOverwriteNonDir = errors.New("cannot overwrite non-directory")
	// ErrReadingSrcDir occurs when attempting to read contents of the source directory fails
	ErrReadingSrcDir = errors.New("cannot read source directory, check source directory permissions")
	// ErrWritingFileToExistingDir occurs when attempting to write a file to an existing directory.
	// See AppendNameToPath option for a more dynamic approach.
	ErrWritingFileToExistingDir = errors.New("cannot overwrite existing directory with file")
	// ErrInvalidBackupControlValue occurs when a control value is given to the Backup option, but the value is invalid.
	ErrInvalidBackupControlValue = errors.New("invalid backup value, valid values are 'off', 'simple', 'existing', 'numbered'")
	// ErrIncompatibleOptions occurs when options are given that are not compatible with one another.
	ErrIncompatibleOptions = errors.New("options given are incompatible")
)
