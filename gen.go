//go:build ignore

package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"golang.org/x/text/unicode/runenames"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	start, count := time.Now(), 0
	for i := rune(0); i < 1_000_000; i++ {
		s := runenames.Name(i)
		if s != "" && !strings.HasPrefix(s, "<") && !slices.ContainsFunc(ignore, func(str string) bool {
			return strings.Contains(s, str)
		}) {
			count++
			fmt.Fprintf(os.Stdout, "%d: %s\n", i, s)
		}
	}
	fmt.Fprintln(os.Stdout, "time:", time.Now().Sub(start).Round(1*time.Microsecond), "count:", count)
	return nil
}

var ignore = []string{
	/*
		"COMPATIBILITY IDEOGRAPH",
		"VARIATION SELECTOR",
		"BLOCK SEXTANT",
	*/
}
