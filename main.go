package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

const maxRedirects = 10

var redirectsFollowed int

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
	locations["ru.aliexpress.com"] = directive{NoQuery: true}
	// locations["www.gearbest.com"] = directive{ParamsToDel: []string{"wid"}}
	locations["www.gearbest.com"] = directive{NoQuery: true}
	locations["www.coolicool.com"] = directive{NoQuery: true}
	locations["www.tinydeal.com"] = directive{NoQuery: true}
	locations["letyshops.ru"] = directive{NoQuery: true, NoPath: true, Scheme: "https"}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		// flag.Usage()
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
		fmt.Println(locParam)
		fmt.Printf("redirectHost = %+v\n", url)
		loc := url.Query().Get(locParam)
		if loc == "" {
			log.Fatalf("%q has no %q param", url.String(), locParam)
		}
		lurl := parseURL(loc)
		fmt.Printf("lurl = %+v\n", lurl)
		if dir, ok := locations[lurl.Host]; ok {
			fmt.Printf("dir = %+v\n", dir)
			fmt.Printf("lurl.Path = %+v\n", lurl.Path)
			fmt.Printf("lurl.RawQuery = %+v\n", lurl.RawQuery)
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
		fmt.Printf("lurl = %+v\n", lurl)
		open.Start(lurl.String())
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
