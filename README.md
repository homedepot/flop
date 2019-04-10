# flop
flop is a Golang file operations library concentrating on safety and feature parity with
[GNU cp](https://www.gnu.org/software/coreutils/manual/html_node/cp-invocation.html).
Most administrators and engineers interact with GNU utilities every day, so it makes sense to utilize
that knowledge and expectations for a library that does the same operations in code.

flop generally tracks cp but strategically diverges from cp where it is advantageous for the programmer to explicitly
define the behavior, like cp assuming that copying from a file path to a directory path means the file should be created
inside the directory. This behavior must be explicitly defined in flop by passing the option AppendNameToPath, otherwise
an error will be returned.

### Usage
Basic file copy.
```go
err := flop.SimpleCopy("src_path", "dst_path")
handle(err)
```

Advanced file copy with options.
```go
options := flop.Options{
    Recursive: true,
    MkdirAll:  true,
}
err := flop.Copy("src_path", "dst_path", options)
handle(err)
```

### Operating Systems Support
flop was built to work on *nix and Windows systems.  The behavior should be consistent across all platforms with the
exception of permissions.  On Windows based systems flop makes no guarantees related to file permissions, ownership,
timestamps, contexts, links, or extended attributes.  These will need to be handled explicitly by the caller, if
applicable, after copy operations are complete.

### Logging
flop won't throw logs at you for no reason, but if you want to follow along with what's going on giving it a logger
can help expose the behavior, or aid in debugging if you are generous enough to contribute.
```go
// the logger just takes a string so format your favorite logger to accept one
import (
	"flop"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	llog "github.com/sirupsen/logrus"
)

func logDebug(msg string) {
	llog.WithFields(llog.Fields{
		"application": "stuffcopy",
	}).Info(msg)
}

func main() {
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	
	err := flop.Copy(src.Name(), dst.Name(), flop.Options{
		InfoLogFunc: zlog.Info().Msg,  // Msg already accepts a string so we can just pass it directly
		DebugLogFunc: logDebug,        // logrus Debug takes ...interface{} so we need to wrap it
	})
	handle(err)
}
```
