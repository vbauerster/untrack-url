// untrack-url
// Copyright (C) 2016-2017 Vladimir Bauer
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/skratchdot/open-golang/open"
	"github.com/vbauerster/untrack-url/tracker"
	"golang.org/x/net/html"
)

const (
	projectHome = "https://github.com/vbauerster/untrack-url"
	cmdName     = "untrack-url"
)

var (
	version = "devel"
	// Command line flags.
	printOnly   bool
	showVersion bool
	// FlagSet
	cmd *flag.FlagSet
)

type directive struct {
	ParamsToDel []string
	NoQuery     bool
	NoPath      bool
	Scheme      string
}

func init() {
	registerTrackers()

	cmd = flag.NewFlagSet(cmdName, flag.ContinueOnError)
	cmd.BoolVar(&printOnly, "p", false, "print only: don't open URL in browser")
	cmd.BoolVar(&showVersion, "v", false, "print version number")

	cmd.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] URL\n\n", cmdName)
		fmt.Println("OPTIONS:")
		cmd.SetOutput(os.Stdout)
		cmd.PrintDefaults()
		fmt.Println()
		fmt.Println("Known trackers:")
		fmt.Println()
		for _, loc := range tracker.KnownTrackers() {
			fmt.Printf("\t%s\n", loc)
		}
		fmt.Println()
		fmt.Printf("project home: %s\n", projectHome)
	}
}

func main() {
	if err := cmd.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if showVersion {
		fmt.Printf("%s: %s (runtime: %s)\n", cmdName, version, runtime.Version())
		os.Exit(0)
	}

	if cmd.NArg() != 1 {
		cmd.Usage()
		os.Exit(2)
	}

	cleanURL, err := tracker.Untrack(cmd.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if printOnly {
		fmt.Println(cleanURL)
	} else if err := open.Start(cleanURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// registerTrackers ...
func registerTrackers() {
	// https://regex101.com/r/kv1rVs/1
	wlocRe := regexp.MustCompile(`(?:window|document)\.location\s*=\s*['"](.*?)['"]`)

	tracker.RegisterTracker("s.click.aliexpress.com", func(tracker *url.URL) (*url.URL, error) {
		// http://ali.ski/gkMqy
		return url.Parse(tracker.Query().Get("dl_target_url"))
	})
	tracker.RegisterTracker("ad.admitad.com", func(tracker *url.URL) (*url.URL, error) {
		// http://fas.st/mKKaRE
		return url.Parse(tracker.Query().Get("ulp"))
	})
	tracker.RegisterTracker("lenkmio.com", func(tracker *url.URL) (*url.URL, error) {
		return url.Parse(tracker.Query().Get("ulp"))
	})
	tracker.RegisterTracker("epnclick.ru", func(tracker *url.URL) (*url.URL, error) {
		// http://ali.pub/2c753s
		return extractEpnRedirect(tracker.String(), wlocRe)
	})
	tracker.RegisterTracker("shopeasy.by", func(tracker *url.URL) (*url.URL, error) {
		// http://ali.pub/2c76pq
		return extractEpnRedirect(tracker.String(), wlocRe)
	})
}

// extracts 'windows.location' value from <script></script> element tag
func extractEpnRedirect(rawurl string, wlocRe *regexp.Regexp) (*url.URL, error) {
	resp, err := http.Get(rawurl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("epn: status not ok")
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	for _, script := range visit(nil, doc) {
		for _, line := range strings.Split(script, "\n") {
			line = strings.Trim(line, " \t")
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			groups := wlocRe.FindStringSubmatch(line)
			if len(groups) > 1 {
				fmt.Println(line)
				fmt.Println(groups[1])
				if targetURL, err := url.Parse(groups[1]); err == nil {
					to := targetURL.Query().Get("to")
					return url.Parse(to)
				}
			}
		}
	}
	return url.Parse(rawurl)
}

func visit(scripts []string, n *html.Node) []string {
	if n.Type == html.ElementNode && n.Data == "script" && n.FirstChild != nil {
		scripts = append(scripts, n.FirstChild.Data)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		scripts = visit(scripts, c)
	}
	return scripts
}
