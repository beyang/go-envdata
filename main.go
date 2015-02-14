package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Config struct {
	// Name of the package to use in generated code. Defaults to 'env'.
	Package string

	// Output defines the output file for the generated code.
	// If left empty, prints to stdout.
	Output string

	// Space separated list of environment variables NOT to capture.
	Ignore string

	// Generate a file that doesn't set any default environment variables.
	Dev bool
}

var alwaysIgnore = []string{"PWD", "SHLVL", "_", "PATH"}

func NewDefaultConfig() *Config {
	return &Config{
		Package: "env",
	}
}

func parseArgs() *Config {
	c := NewDefaultConfig()
	flag.StringVar(&c.Package, "pkg", c.Package, "Package name to use in generated code.")
	flag.StringVar(&c.Output, "o", c.Output, "Optional name of the output file to be generated.")
	flag.StringVar(&c.Ignore, "ignore", c.Ignore, "Space separated list of environment variables to ignore.")
	flag.BoolVar(&c.Dev, "dev", c.Dev, "Do not capture the environment, but instead generate a file that just reads from the runtime environment.")
	flag.Parse()

	return c
}

func writeDev(w io.Writer, c *Config) error {
	return writeRelease(w, c, nil)
}

func writeRelease(w io.Writer, c *Config, env map[string]string) error {
	var defaultsMapExpr string
	if len(env) == 0 {
		defaultsMapExpr = ""
	} else {
		var kvExprs []string
		for k, v := range env {
			kvExprs = append(kvExprs, fmt.Sprintf("\t%q: %q,", k, v))
		}
		sort.Strings(kvExprs) // sort for stable ordering
		defaultsMapExpr = "\n" + strings.Join(kvExprs, "\n") + "\n"
	}

	_, err := fmt.Fprintf(w, `// Package %s sets default values for environment variables.
// Usage: in any package that calls os.Getenv or references the environment, include:
//
//     import _ "full/path/to/%s"
//
package %s

import "os"

var defaults = map[string]string{%s}

func init() {
	setDefaultEnv()
}

func setDefaultEnv() {
	for k, v := range defaults {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}
`, c.Package, c.Package, c.Package, defaultsMapExpr)
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
	for _, ig := range alwaysIgnore {
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
	var out io.Writer
	if c.Output == "" {
		out = os.Stdout
	} else {
		fd, err := os.Create(c.Output)
		if err != nil {
			return nil
		}
		defer fd.Close()
		bfd := bufio.NewWriter(fd)
		defer bfd.Flush()
		out = bfd
	}

	if c.Dev {
		writeDev(out, c)
	} else {
		writeRelease(out, c, env)
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
