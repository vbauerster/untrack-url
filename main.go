// untrack-url
// Copyright (C) 2016-2017 Vladimir Bauer
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
	"github.com/vbauerster/untrack-url/ranger"
)

const (
	projectHome = "https://github.com/vbauerster/untrack-url"
	cmdName     = "untrack-url"
)

var (
	version = "devel"
	// Command line flags.
	printOnly   bool
	debug       bool
	showVersion bool
	// FlagSet
	cmd *flag.FlagSet
)

func init() {
	// NewFlagSet for the sake of cmd.SetOutput
	cmd = flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmd.SetOutput(os.Stdout)
	cmd.BoolVar(&printOnly, "p", false, "print only: don't open URL in browser")
	cmd.BoolVar(&debug, "d", false, "debug: print debug info, implies -p")
	cmd.BoolVar(&showVersion, "v", false, "print version number")

	cmd.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] URL\n\n", cmdName)
		fmt.Println("OPTIONS:")
		cmd.PrintDefaults()
		fmt.Println()
		fmt.Println("Known trackers:")
		fmt.Println()
		for _, host := range ranger.KnownTrackers() {
			fmt.Printf("\t%s\n", host)
		}
		fmt.Println()
		fmt.Println("Known shops:")
		fmt.Println()
		for _, host := range ranger.KnownShops() {
			fmt.Printf("\t%s\n", host)
		}
		fmt.Println()
		fmt.Printf("Project home: %s\n", projectHome)
	}
}

func main() {
	cmd.Parse(os.Args[1:])

	if showVersion {
		fmt.Printf("%s: %s (runtime: %s)\n", cmdName, version, runtime.Version())
		os.Exit(0)
	}

	if cmd.NArg() != 1 {
		cmd.Usage()
		os.Exit(2)
	}

	ranger.Debug = debug
	cleanURL, err := ranger.Untrack(cmd.Arg(0))
	if err != nil {
		if _, ok := errors.Cause(err).(ranger.UntrackErr); ok {
			if debug {
				fmt.Fprintf(os.Stderr, "%+v\n", err)
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			fmt.Println("There was an unexpected error; please report this as a bug.")
		}
		os.Exit(1)
	}

	if printOnly || debug {
		fmt.Println(cleanURL)
	} else if err := open.Start(cleanURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
