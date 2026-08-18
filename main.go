package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dir, port := parseFlags(os.Args[1:])
	if dir == "" {
		fmt.Fprintln(os.Stderr, "usage: mdwiki <folder> [port] (or: mdwiki -dir <folder> -port <port>)")
		os.Exit(1)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("resolving folder: %v", err)
	}
	if fi, err := os.Stat(absDir); err != nil || !fi.IsDir() {
		log.Fatalf("not a directory: %s", absDir)
	}

	if err := serve(absDir, port); err != nil {
		log.Fatal(err)
	}
}

func parseFlags(args []string) (string, int) {
	fs := flag.NewFlagSet("mdwiki", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	dirFlag := fs.String("dir", "", "path to the docs folder (root of the markdown wiki)")
	portFlag := fs.Int("port", 8888, "port to listen on")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	dir := *dirFlag
	port := *portFlag

	positional := fs.Args()
	if dir == "" && len(positional) >= 1 {
		dir = positional[0]
	}
	if len(positional) >= 2 {
		_, _ = fmt.Sscanf(positional[1], "%d", &port)
	}
	return dir, port
}
