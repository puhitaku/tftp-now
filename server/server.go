package server

import (
	"errors"
	"io"
	"os"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var root string

func SetRoot(p string) {
	root = p
}

// ReadHandler is called when client starts file download from server
func ReadHandler(requestedPath string, rf io.ReaderFrom) error {
	reqID := ulid.Make().String()
	log.Info().Str("requestId", reqID).Msgf("read request: %s", requestedPath)

	evalPath, err := evaluatePath(requestedPath, root, true)
	if err != nil {
		perr := InvalidPathError{}
		if errors.As(err, &perr) {
			log.Error().Str("requestId", reqID).EmbedObject(perr).Msgf("failed to evaluate path")
		} else {
			log.Error().Str("requestId", reqID).Msgf("failed to evaluate path: %s", err)
		}
		return err
	}

	log.Debug().Str("requestId", reqID).Msgf("evaluated path: %s", evalPath)

	file, err := os.Open(evalPath)
	if err != nil {
		log.Error().Str("requestId", reqID).Msgf("failed to open the file: %s", err)
		return err
	}
	defer file.Close()

	n, err := rf.ReadFrom(file)
	if err != nil {
		log.Error().Str("requestId", reqID).Msgf("failed to read from the file: %s", err)
		return err
	}
	log.Info().Str("requestId", reqID).Int64("bytes", n).Msg("successfully handled")
	return nil
}

// WriteHandler is called when client starts file upload to server
func WriteHandler(requestedPath string, wt io.WriterTo) error {
	reqID := ulid.Make().String()
	log.Info().Str("requestId", reqID).Msgf("write request: %s", requestedPath)

	evalPath, err := evaluatePath(requestedPath, root, false)
	if err != nil {
		perr := InvalidPathError{}
		if errors.As(err, &perr) {
			log.Error().Str("requestId", reqID).EmbedObject(perr).Msgf("failed to evaluate path")
		} else {
			log.Error().Str("requestId", reqID).Msgf("failed to evaluate path: %s", err)
		}
		return err
	}

	log.Debug().Str("requestId", reqID).Msgf("evaluated path: %s", evalPath)

	file, err := os.OpenFile(evalPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Error().Str("requestId", reqID).Msgf("failed to open the file: %s", err)
		return err
	}
	defer file.Close()

	n, err := wt.WriteTo(file)
	if err != nil {
		log.Error().Str("requestId", reqID).Msgf("failed to write to the file: %s", err)
		return err
	}
	log.Info().Str("requestId", reqID).Int64("bytes", n).Msg("successfully handled")
	return nil
}
