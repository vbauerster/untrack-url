package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

const maxRedirects = 10

var (
	// Command line flags.
	printOnly   bool
	debug       bool
	showVersion bool

	// number of redirects followed
	redirectsFollowed int

	version     = "0.0.1"
	projectHome = "https://github.com/vbauerster/radali"
)

type directive struct {
	ParamsToDel []string
	NoQuery     bool
	NoPath      bool
	Scheme      string
}

var redirectHosts = map[string]string{
	"s.click.aliexpress.com": "dl_target_url",
	"ad.admitad.com":         "ulp",
}

var locations = make(map[string]directive)

func init() {
	flag.BoolVar(&printOnly, "p", false, "print only: don't open URL in browser")
	flag.BoolVar(&debug, "d", false, "debug: print debug info")
	flag.BoolVar(&showVersion, "v", false, "print version number")

	locations["ru.aliexpress.com"] = directive{NoQuery: true}
	// locations["www.gearbest.com"] = directive{ParamsToDel: []string{"wid"}}
	locations["www.gearbest.com"] = directive{NoQuery: true}
	locations["www.coolicool.com"] = directive{NoQuery: true}
	locations["www.tinydeal.com"] = directive{NoQuery: true}
	locations["letyshops.ru"] = directive{NoQuery: true, NoPath: true, Scheme: "https"}

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: radali [OPTIONS] URL\n\n")
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "project home: %s\n", projectHome)
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("radali %s (runtime: %s)\n", version, runtime.Version())
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(2)
	}

	url := parseURL(args[0])
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	visit(client, url)
}

func visit(client *http.Client, url *url.URL) {
	resp, err := client.Get(url.String())
	if err != nil {
		log.Fatalf("failed to read response: %v", err)
	}
	resp.Body.Close()

	if locParam, ok := redirectHosts[url.Host]; ok {
		loc := url.Query().Get(locParam)
		if loc == "" {
			log.Fatalf("%q has no %q param", url.String(), locParam)
		}
		lurl := parseURL(loc)
		if debug {
			fmt.Printf("redirectHost = %+v\n", url)
			fmt.Printf("%s = %+v\n", locParam, lurl)
		}
		if dir, ok := locations[lurl.Host]; ok {
			if debug {
				fmt.Printf("dir = %+v\n", dir)
				fmt.Printf("%s.Path = %+v\n", locParam, lurl.Path)
				fmt.Printf("%s.RawQuery = %+v\n", locParam, lurl.RawQuery)
			}
			if dir.NoQuery {
				lurl.RawQuery = ""
			} else if len(dir.ParamsToDel) != 0 {
				v := lurl.Query()
				for _, param := range dir.ParamsToDel {
					v.Del(param)
				}
				lurl.RawQuery = v.Encode()
			}
			if dir.NoPath {
				lurl.Path = ""
			}
			if dir.Scheme != "" {
				lurl.Scheme = dir.Scheme
			}
		}
		if printOnly || debug {
			fmt.Println(lurl)
		} else {
			open.Start(lurl.String())
		}
		return
	}

	if isRedirect(resp.StatusCode) {
		loc, err := resp.Location()
		if err != nil {
			if err == http.ErrNoLocation {
				// 30x but no Location to follow, give up.
				return
			}
			log.Fatalf("unable to follow redirect: %v", err)
		}

		redirectsFollowed++
		if redirectsFollowed > maxRedirects {
			log.Fatalf("maximum number of redirects (%d) followed", maxRedirects)
		}

		visit(client, loc)
	}
}

func parseURL(uri string) *url.URL {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	url, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("could not parse url %q: %v", uri, err)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
		if !strings.HasSuffix(url.Host, ":80") {
			url.Scheme += "s"
		}
	}
	return url
}

func isRedirect(status int) bool {
	return status > 299 && status < 400
}
