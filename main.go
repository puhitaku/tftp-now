package main

import (
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

var usage = `
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
  $ tftp-now read -host 192.168.1.1 -path dir/foo -output bar

Example (write): send 'bar' to '{server root}/dir/foo' of 192.168.1.1.
  $ tftp-now write -host 192.168.1.1 -path dir/foo -input bar
`

func main() {
	os.Exit(main_())
}

func main_() int {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	if len(os.Args) < 2 {
		fmt.Print(usage)
		return 1
	}

	switch os.Args[1] {
	case "serve":
		serverCmd := flag.NewFlagSet("tftp-now serve [<options>]", flag.ExitOnError)
		host := serverCmd.String("host", "0.0.0.0", "Host address")
		port := serverCmd.Int("port", 69, "Port number")
		root := serverCmd.String("root", ".", "Directory path")

		err := serverCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic().Msgf("failed to parse args: %s", err)
		}

		abs, err := filepath.Abs(*root)
		if err != nil {
			log.Panic().Msgf("failed to get the absolute path: %s", err)
		}

		server.SetRoot(abs)
		s := tftp.NewServer(server.ReadHandler, server.WriteHandler)
		s.SetTimeout(5 * time.Second)

		log.Info().Str("host", *host).Int("port", *port).Str("directory", abs).Msg("TFTP server is up")
		err = s.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port))
		if err != nil {
			log.Err(err)
			return 1
		}
	case "read":
		clientCmd := flag.NewFlagSet("tftp-now read [<options>]", flag.ExitOnError)
		host := clientCmd.String("host", "127.0.0.1", "Host address")
		port := clientCmd.Int("port", 69, "Port number")
		path := clientCmd.String("path", "", "Remote file path to read from (REQUIRED)")
		output := clientCmd.String("output", "", "Local file path to save to (REQUIRED)")

		if len(os.Args) < 2 {
			clientCmd.Usage()
			return 1
		}

		err := clientCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic().Msgf("failed to parse args: %s", err)
		}

		if *path == "" {
			log.Fatal().Msgf("please specify '-path'")
			return 1
		} else if *output == "" {
			log.Fatal().Msgf("please specify '-output'")
			return 1
		}

		file, err := os.Create(*output)
		if err != nil {
			log.Error().Msgf("failed to open '%s' to write: %s", *output, err)
			return 1
		}
		defer file.Close()

		cli, err := tftp.NewClient(fmt.Sprintf("%s:%d", *host, *port))
		if err != nil {
			log.Error().Msgf("failed to create a new client: %s", err)
			return 1
		}

		tf, err := cli.Receive(*path, "octet")
		if err != nil {
			log.Error().Msgf("failed to receive '%s': %s", *path, err)
			return 1
		}

		n, err := tf.WriteTo(file)
		if err != nil {
			log.Error().Msgf("failed to write the received data to '%s': %s", *output, err)
			return 1
		}

		log.Info().Int64("length", n).Msgf("successfully received")
	case "write":
		clientCmd := flag.NewFlagSet("tftp-now write [<options>]", flag.ExitOnError)
		host := clientCmd.String("host", "127.0.0.1", "Host address")
		port := clientCmd.Int("port", 69, "Port number")
		path := clientCmd.String("path", "", "Remote file path to save to (REQUIRED)")
		input := clientCmd.String("input", "", "Local file path to read from (REQUIRED")

		if len(os.Args) < 2 {
			clientCmd.Usage()
			return 1
		}

		err := clientCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic().Msgf("failed to parse args: %s", err)
		}

		file, err := os.Open(*input)
		if err != nil {
			log.Error().Msgf("failed to open '%s' to write: %s", *input, err)
			return 1
		}
		defer file.Close()

		cli, err := tftp.NewClient(fmt.Sprintf("%s:%d", *host, *port))
		if err != nil {
			log.Error().Msgf("failed to create a new client: %s", err)
			return 1
		}

		rf, err := cli.Send(*path, "octet")
		if err != nil {
			log.Error().Msgf("failed to send '%s': %s", *path, err)
			return 1
		}

		n, err := rf.ReadFrom(file)
		if err != nil {
			log.Error().Msgf("failed to read the sending data from '%s': %s", *input, err)
			return 1
		}

		log.Info().Int64("length", n).Msgf("successfully sent")
	default:
		fmt.Println("Invalid command. Use 'serve', 'read', or 'write'")
		return 1
	}

	return 0
}
