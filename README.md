# tftp-now

*Single-binary TFTP server and client that you can use right now. No package installation, no configuration, no frustration.*


# tl;dr

1. Download the latest executable.

    - Linux [amd64](https://github.com/puhitaku/tftp-now/releases/latest/download/tftp-now-linux-amd64) /
            [arm](https://github.com/puhitaku/tftp-now/releases/latest/download/tftp-now-linux-arm) /
            [arm64](https://github.com/puhitaku/tftp-now/releases/latest/download/tftp-now-linux-arm64) /
            [riscv64](https://github.com/puhitaku/tftp-now/releases/latest/download/tftp-now-linux-riscv64)
    - macOS [amd64](https://github.com/puhitaku/tftp-now/releases/latest/download/tftp-now-darwin-amd64) /
            [arm64](https://github.com/puhitaku/tftp-now/releases/latest/download/tftp-now-darwin-arm64)
    - Windows [amd64](https://github.com/puhitaku/tftp-now/releases/latest/download/tftp-now-windows-amd64.exe) /
              [arm64](https://github.com/puhitaku/tftp-now/releases/latest/download/tftp-now-windows-arm64.exe)

1. Make it executable.

   ```
   $ chmod +x tftp-now-darwin-arm64  # example for macOS
   ```

1. Run it.

    1. As a server: `tftp-now-{OS}-{ARCH} serve`
    1. As a client, to read (receive): `tftp-now-{OS}-{ARCH} read -remote remote/path/to/read.bin -local read.bin`
    1. As a client, to write (send): `tftp-now-{OS}-{ARCH} write -remote remote/path/to/write.bin -local write.bin`


# Why

I enjoy installing OpenWrt onto routers, but one of the main challenges is transferring it to the bootloader via TFTP. To do this, a temporary TFTP server or client is necessary, but I have always struggled with setting up a TFTP server.

While macOS has an out-of-the-box TFTP server, it requires running launchctl to invoke the hidden server. The process is tricky, and I always Google for guidance.

Linux distros, on the other hand, usually don't have a built-in TFTP server. Installing tftpd via apt is an option, but it's configured for inetd by default and requires some additional configuration. Only the manpage and Google can provide guidance on how to do it properly.

As for Windows, it doesn't come with a TFTP server by default, except for the server variants. Community-based software is the first choice for Windows, and again, Google is the go-to source for finding the right software to download.

It's frustrating that setting up a TFTP server is always such a hassle. This is why I created tftp-now.


# How

Fortunately, there is a well-developed TFTP server/client implementation for Golang available at https://github.com/pin/tftp. The example code snippet provided on the site is exactly what I was looking for. However, since it's just an example, it lacks proper security checks and validation. To address this, I implemented these features myself and integrated the package into a simple CLI.


# Download & Run

Download the latest executable from [the release page](https://github.com/puhitaku/tftp-now/releases/latest). If there's no binary that runs on your system, please raise an issue.

```
$ tftp-now

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
```

