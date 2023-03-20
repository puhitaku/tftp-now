package server

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var invalidPathError = errors.New("invalid path, or not found")

func evaluatePath(untrustedPath, trustedRoot string, evalSymlink bool) (string, error) {
	joined := filepath.Join(trustedRoot, untrustedPath) // Join() also cleans the path
	abs, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("failed to get the absolute path: %w", err)
	}

	evaluated := abs

	if evalSymlink {
		evaluated, err = filepath.EvalSymlinks(abs)
		if err != nil {
			return "", invalidPathError
		}
	}

	trusted := strings.HasPrefix(evaluated, trustedRoot)
	if !trusted {
		return "", invalidPathError
	}

	return evaluated, nil
}
