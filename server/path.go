package server

import (
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

type InvalidPathError struct {
	PathRequest string
	PathAbs     string
	PathEval    string
	Reason      reason

	BaseErr error
}

type reason string

const (
	AbsFail   reason = "failed to get the absolute path"
	EvalFail  reason = "failed to evaluate symlinks"
	OutOfRoot reason = "out of the root directory"
)

func (i InvalidPathError) Error() string {
	// Prevent unexpected exposure of the actual reason, for security
	return "invalid path, or not found"
}

func (i InvalidPathError) Unwrap() error {
	return i.BaseErr
}

func (i InvalidPathError) MarshalZerologObject(e *zerolog.Event) {
	e.Str("pathRequest", i.PathRequest)
	if i.PathAbs != "" {
		e.Str("pathAbs", i.PathAbs)
	}
	if i.PathEval != "" {
		e.Str("pathEval", i.PathEval)
	}
	e.Str("reason", string(i.Reason))
	e.AnErr("baseErr", i.BaseErr)
}

func evaluatePath(requestPath, trustedRoot string, evalSymlink bool) (string, error) {
	joined := filepath.Join(trustedRoot, requestPath) // Join() also cleans the path
	abs, err := filepath.Abs(joined)
	if err != nil {
		return "", InvalidPathError{
			PathRequest: requestPath,
			Reason:      AbsFail,
			BaseErr:     err,
		}
	}

	evaluated := abs

	if evalSymlink {
		evaluated, err = filepath.EvalSymlinks(abs)
		if err != nil {
			return "", InvalidPathError{
				PathRequest: requestPath,
				PathAbs:     abs,
				Reason:      EvalFail,
				BaseErr:     err,
			}
		}
	}

	trusted := strings.HasPrefix(evaluated, trustedRoot)
	if !trusted {
		return "", InvalidPathError{
			PathRequest: requestPath,
			PathAbs:     abs,
			PathEval:    evaluated,
			Reason:      OutOfRoot,
			BaseErr:     err,
		}
	}

	return evaluated, nil
}
