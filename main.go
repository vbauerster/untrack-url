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

	target := removeAds(follow(args[0]))
	if printOnly || debug {
		fmt.Println(target)
	} else {
		open.Start(target)
	}
}

func follow(url string) string {
	// number of redirects followed
	var redirectsFollowed int
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	next := parseURL(url)
	for {
		resp, err := client.Get(next.String())
		if err != nil {
			log.Fatalf("failed to read response: %v", err)
		}
		resp.Body.Close()

		if isRedirect(resp.StatusCode) {
			loc, err := resp.Location()
			if err != nil {
				log.Fatalf("unable to follow redirect: %v", err)
			}

			if p, ok := redirectHosts[next.Host]; ok {
				if _, ok = next.Query()[p]; ok {
					if debug {
						fmt.Printf("found ref: %q\n", loc)
					}
					return loc.String()
				}
			}

			redirectsFollowed++
			if redirectsFollowed > maxRedirects {
				log.Fatalf("maximum number of redirects (%d) followed", maxRedirects)
			}
			next = loc
		} else {
			break
		}
	}
	return ""
}

func removeAds(ref string) string {
	url := parseURL(ref)
	if dir, ok := locations[url.Host]; ok {
		if debug {
			fmt.Printf("%s = %+v\n", url.Host, dir)
		}
		if dir.NoQuery {
			url.RawQuery = ""
		} else if len(dir.ParamsToDel) != 0 {
			v := url.Query()
			for _, param := range dir.ParamsToDel {
				v.Del(param)
			}
			url.RawQuery = v.Encode()
		}
		if dir.NoPath {
			url.Path = ""
		}
		if dir.Scheme != "" {
			url.Scheme = dir.Scheme
		}
	}
	return url.String()
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
