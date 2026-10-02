package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	Version   = "dev"
	GitCommit = ""
	BuildTime = ""
)

type cliOptions struct {
	dir         string
	port        int
	showVersion bool
}

func main() {
	options, err := parseFlags(os.Args[1:])
	if err != nil {
		os.Exit(2)
	}
	if options.showVersion {
		fmt.Println(VersionInfo())
		return
	}

	if options.dir == "" {
		fmt.Fprintln(os.Stderr, "usage: mdwiki <folder> [port] (or: mdwiki -dir <folder> -port <port>)")
		os.Exit(1)
	}

	absDir, err := filepath.Abs(options.dir)
	if err != nil {
		log.Fatalf("resolving folder: %v", err)
	}
	if fi, err := os.Stat(absDir); err != nil || !fi.IsDir() {
		log.Fatalf("not a directory: %s", absDir)
	}

	if err := serve(absDir, options.port); err != nil {
		log.Fatal(err)
	}
}

func parseFlags(args []string) (cliOptions, error) {
	fs := flag.NewFlagSet("mdwiki", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	dirFlag := fs.String("dir", "", "path to the docs folder (root of the markdown wiki)")
	portFlag := fs.Int("port", 8888, "port to listen on")
	versionFlag := fs.Bool("version", false, "Show version information and exit")

	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}

	options := cliOptions{
		dir:         *dirFlag,
		port:        *portFlag,
		showVersion: *versionFlag,
	}

	positional := fs.Args()
	if options.dir == "" && len(positional) >= 1 {
		options.dir = positional[0]
	}
	if len(positional) >= 2 {
		_, _ = fmt.Sscanf(positional[1], "%d", &options.port)
	}
	return options, nil
}

func VersionInfo() string {
	parts := []string{"mdwiki " + Version}
	if GitCommit != "" {
		parts = append(parts, "commit "+GitCommit)
	}
	if BuildTime != "" {
		parts = append(parts, "built "+BuildTime)
	}
	return strings.Join(parts, ", ")
}
