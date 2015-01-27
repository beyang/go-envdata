package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Config struct {
	// Name of the package to use. Defaults to 'main'.
	Package string

	// Output defines the output file for the generated code.
	// If left empty, this defaults to 'envdata.go' in the current
	// working directory.
	Output string

	// Space separated list of environment variables NOT to capture. Defaults to
	// some always-set ones like PWD that you probably don't want captured.
	Ignore string

	// Generate a file that doesn't contain the current environment, but instead
	// just returns the runtime environment variable values.
	//
	// Using this file is equivalent to using os.Getenv, but it conveniently has
	// the same interface as the file generated in release mode, so your code
	// doesn't have to change.
	Dev bool
}

func NewDefaultConfig() *Config {
	return &Config{
		Package: "main",
		Output:  "./envdata.go",
		Ignore:  "PWD SHLVL _",
	}
}

func parseArgs() *Config {
	c := NewDefaultConfig()
	flag.StringVar(&c.Package, "pkg", c.Package, "Package name to use in generated code.")
	flag.StringVar(&c.Output, "o", c.Output, "Optional name of the output file to be generated.")
	flag.StringVar(&c.Ignore, "ignore", c.Ignore, "Space separated list of environment variables to ignore.")
	flag.BoolVar(&c.Dev, "dev", c.Dev, "Do not capture the environment, but instead generate a file that just reads from the runtime environment.")
	return c
}

func writeDev(w io.Writer) error {
	return writeRelease(w, nil)
}

func writeRelease(w io.Writer, env map[string]string) error {
	var defaultsMapExpr string
	if len(env) == 0 {
		defaultsMapExpr = ""
	} else {
		var kvExprs []string
		for k, v := range env {
			kvExprs = append(kvExprs, fmt.Sprintf("\t%q: %q,", k, v))
		}
		defaultsMapExpr = "\n" + strings.Join(kvExprs, "\n") + "\n"
	}

	_, err := fmt.Fprintf(w, `package env

import "os"

var defaults = map[string]string{%s}

// Env returns the value of an environment variable if set. Otherwise, it returns the default value if it exists.
func Env(name string) string {
    if value := os.Getenv(name); value != "" {
        return value
    }
    return defaults[name]
}
`, defaultsMapExpr)
	if err != nil {
		return err
	}
	return nil
}

// Transcribe reads environment variables and emits a Go source file with their default values.
func Transcribe(c *Config) error {
	// Get environment to capture
	ignore_ := strings.Fields(c.Ignore)
	ignore := make(map[string]bool)
	for _, ig := range ignore_ {
		ignore[ig] = true
	}
	env := make(map[string]string)
	for _, keyval := range os.Environ() {
		key := strings.SplitN(keyval, "=", 2)[0]
		val := os.Getenv(key)
		if !ignore[key] {
			env[key] = val
		}
	}

	// Write file
	fd, err := os.Create(c.Output)
	if err != nil {
		return nil
	}
	defer fd.Close()

	bfd := bufio.NewWriter(fd)
	defer bfd.Flush()

	if c.Dev {
		writeDev(bfd)
	} else {
		writeRelease(bfd, env)
	}
	return nil
}

func main() {
	c := parseArgs()
	if err := Transcribe(c); err != nil {
		fmt.Fprintf(os.Stderr, "go-envdata: %v\n", err)
		os.Exit(1)
	}
}
