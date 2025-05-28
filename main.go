package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pin/tftp/v3"
	"github.com/puhitaku/tftp-now/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed VERSION
var version string

var usage = fmt.Sprintf(`
tftp-now %s

Usage of tftp-now:
  tftp-now <command> [<args>]

Server Commands:
  serve  Start TFTP server

Client Commands:
  read   Read a file from a TFTP server
  write  Write a file to a TFTP server

Other Commands:
  help   Show this help


Example (serve): start serving on 0.0.0.0:69
  $ tftp-now serve

Example (read): receive '{server root}/dir/foo' from 192.168.1.1 and save it to 'bar'.
  $ tftp-now read -host 192.168.1.1 -remote dir/foo -local bar

Example (write): send 'bar' to '{server root}/dir/foo' of 192.168.1.1.
  $ tftp-now write -host 192.168.1.1 -remote dir/foo -local bar


Tips:
  - If tftp-now executable itself or a link to tftp-now is named "tftp-now-serve",
    tftp-now will start a TFTP server without any explicit subcommand. Please specify
    a subcommand if you want to specify options.
  - The block size for the server will be clamped to the smaller of the block size
    a client requests and the MTU (minus overhead) of the interface.
`, version)[1:]

func main() {
	os.Exit(main_())
}

func main_() int {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(zerolog.InfoLevel)

	const (
		serve = "serve"
		read  = "read"
		write = "write"
		help  = "help"
	)

	command := help
	options := []string{}

	if len(os.Args) > 1 {
		command = os.Args[1]
		options = os.Args[2:]
	} else if filepath.Base(os.Args[0]) == "tftp-now-serve" {
		log.Info().Msgf("tftp-now will start a server since the executable's name is 'tftp-now-serve'")
		command = serve
	}

	// validateBlockSize validates the block size.
	// While RFC 2348 defines that block size must be between 8 and 65464,
	// pin/tftp will not accept and silently ignores the length less than 512 Bytes.
	validateBlockSize := func(l int) (valid bool) {
		valid = 512 <= l && l <= 65464
		if !valid {
			log.Error().Msgf("block size must be between 512 and 65464")
		}
		return
	}

	switch command {
	case serve:
		serverCmd := flag.NewFlagSet("tftp-now serve [<options>]", flag.ExitOnError)
		host := serverCmd.String("host", "0.0.0.0", "Host address")
		port := serverCmd.Int("port", 69, "Port number")
		root := serverCmd.String("root", ".", "Root directory path")
		blkSize := serverCmd.Int("blksize", 512, "Block size")
		verbose := serverCmd.Bool("verbose", false, "Enable verbose debug output")

		err := serverCmd.Parse(options)
		if err != nil {
			log.Error().Msgf("failed to parse args: %s", err)
			return 1
		}

		if !validateBlockSize(*blkSize) {
			return 1
		}

		if *verbose {
			log.Logger = log.Logger.Level(zerolog.DebugLevel)
		}

		abs, err := filepath.Abs(*root)
		if err != nil {
			log.Error().Msgf("failed to get the absolute path: %s", err)
			return 1
		}

		server.SetRoot(abs)
		s := tftp.NewServer(server.ReadHandler, server.WriteHandler)
		s.SetBlockSize(*blkSize)
		s.SetTimeout(5 * time.Second)

		log.Info().Str("host", *host).Int("port", *port).Str("directory", abs).Int("blocksize", *blkSize).Msg("starting the TFTP server")
		err = s.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port))
		if err != nil {
			log.Error().Msgf("failed to run the server: %s", err)
			return 1
		}
	case read:
		clientCmd := flag.NewFlagSet("tftp-now read [<options>]", flag.ExitOnError)
		host := clientCmd.String("host", "127.0.0.1", "Host address")
		port := clientCmd.Int("port", 69, "Port number")
		remote := clientCmd.String("remote", "", "Remote file path to read from (REQUIRED)")
		local := clientCmd.String("local", "", "Local file path to save to (if unspecified, inferred from -remote)")
		blkSize := clientCmd.Int("blksize", 512, "Block size")

		if len(options) < 2 {
			clientCmd.Usage()
			return 1
		}

		err := clientCmd.Parse(options)
		if err != nil {
			log.Error().Msgf("failed to parse args: %s", err)
			return 1
		}

		if *remote == "" {
			log.Error().Msgf("please specify '-remote'")
			return 1
		}

		if *local == "" {
			*local = filepath.Base(*remote)
		}

		if !validateBlockSize(*blkSize) {
			return 1
		}

		log.Info().Str("host", fmt.Sprintf("%s:%d", *host, *port)).Str("remote", *remote).Str("local", *local).Int("blocksize", *blkSize).Msgf("start reading")

		cli, err := tftp.NewClient(fmt.Sprintf("%s:%d", *host, *port))
		if err != nil {
			log.Error().Msgf("failed to create a new client: %s", err)
			return 1
		}

		cli.SetBlockSize(*blkSize)

		tf, err := cli.Receive(*remote, "octet")
		if err != nil {
			log.Error().Msgf("failed to receive '%s': %s", *remote, err)
			return 1
		}

		file, err := os.Create(*local)
		if err != nil {
			log.Error().Msg(err.Error())
			return 1
		}
		defer file.Close()

		n, err := tf.WriteTo(file)
		if err != nil {
			log.Error().Msgf("failed to write the received data to '%s': %s", *local, err)
			return 1
		}

		log.Info().Int64("length", n).Msgf("successfully received")
	case write:
		clientCmd := flag.NewFlagSet("tftp-now write [<options>]", flag.ExitOnError)
		host := clientCmd.String("host", "127.0.0.1", "Host address")
		port := clientCmd.Int("port", 69, "Port number")
		remote := clientCmd.String("remote", "", "Remote file path to save to (REQUIRED)")
		local := clientCmd.String("local", "", "Local file path to read from (REQUIRED)")
		blkSize := clientCmd.Int("blksize", 512, "Block size")

		if len(options) < 2 {
			clientCmd.Usage()
			return 1
		}

		err := clientCmd.Parse(options)
		if err != nil {
			log.Error().Msgf("failed to parse args: %s", err)
			return 1
		}

		if *remote == "" {
			log.Error().Msgf("please specify '-remote'")
			return 1
		} else if *local == "" {
			log.Error().Msgf("please specify '-local'")
			return 1
		}

		file, err := os.Open(*local)
		if err != nil {
			log.Error().Msg(err.Error())
			return 1
		}
		defer file.Close()

		if !validateBlockSize(*blkSize) {
			return 1
		}

		log.Info().Str("host", fmt.Sprintf("%s:%d", *host, *port)).Str("remote", *remote).Str("local", *local).Int("blocksize", *blkSize).Msgf("start writing")

		cli, err := tftp.NewClient(fmt.Sprintf("%s:%d", *host, *port))
		if err != nil {
			log.Error().Msgf("failed to create a new client: %s", err)
			return 1
		}

		cli.SetBlockSize(*blkSize)

		rf, err := cli.Send(*remote, "octet")
		if err != nil {
			log.Error().Msgf("failed to send '%s': %s", *remote, err)
			return 1
		}

		n, err := rf.ReadFrom(file)
		if err != nil {
			log.Error().Msgf("failed to read the sending data from '%s': %s", *local, err)
			return 1
		}

		log.Info().Int64("length", n).Msgf("successfully sent")
	case help:
		fmt.Print(usage)
		return 1
	default:
		fmt.Println("Invalid command. Specify 'serve', 'read', 'write', or 'help'.")
		return 1
	}

	return 0
}
