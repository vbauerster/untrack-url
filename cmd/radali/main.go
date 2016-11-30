package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"golang.org/x/net/html"

	"github.com/skratchdot/open-golang/open"
)

const (
	maxRedirects = 10
	projectHome  = "https://github.com/vbauerster/radali"
	cmdName      = "radali"
)

var (
	version = "devel"
	// Command line flags.
	printOnly   bool
	debug       bool
	showVersion bool
	// FlagSet
	cmd    *flag.FlagSet
	wlocRe *regexp.Regexp
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
	"lenkmio.com":            "ulp",
	"cashback.epn.bz":        "inviter",
	"alibonus.com":           "u",
}

var locations = make(map[string]directive)

func init() {
	wlocRe = regexp.MustCompile(`(?:window\.)?location\s*=\s*['"](.*?)['"]`)

	cmd = flag.NewFlagSet(cmdName, flag.ContinueOnError)
	cmd.BoolVar(&printOnly, "p", false, "print only: don't open URL in browser")
	cmd.BoolVar(&debug, "d", false, "debug: print debug info")
	cmd.BoolVar(&showVersion, "v", false, "print version number")

	locations["ru.aliexpress.com"] = directive{NoQuery: true}
	locations["www.gearbest.com"] = directive{NoQuery: true}
	locations["www.coolicool.com"] = directive{NoQuery: true}
	locations["www.tinydeal.com"] = directive{NoQuery: true}
	locations["www.banggood.com"] = directive{NoQuery: true}
	locations["letyshops.ru"] = directive{NoQuery: true, NoPath: true, Scheme: "https"}
	locations["cashback.epn.bz"] = directive{NoQuery: true, NoPath: true}
	locations["alibonus.com"] = directive{NoQuery: true, NoPath: true}

	cmd.Usage = usage
}

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] URL\n\n", cmdName)
	fmt.Println("OPTIONS:")
	cmd.SetOutput(os.Stdout)
	cmd.PrintDefaults()
	fmt.Println()
	fmt.Println("Supported resources:")
	fmt.Println()
	for _, loc := range supportedLocations() {
		fmt.Printf("\t%s\n", loc)
	}
	fmt.Println()
	fmt.Printf("project home: %s\n", projectHome)
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
		fmt.Printf("%s %s (runtime: %s)\n", cmdName, version, runtime.Version())
		os.Exit(0)
	}

	args := cmd.Args()
	if len(args) != 1 {
		cmd.Usage()
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

			// Special case for "epnclick.ru"
			if loc.Host == "epnclick.ru" {
				if debug {
					fmt.Printf("fetching: %q\n", loc.String())
				}
				return extractEpnRedirect(loc.String())
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
	if ref == "" {
		fmt.Println("Nothing found!")
		fmt.Println("Supported resources:")
		fmt.Println()
		for _, loc := range supportedLocations() {
			fmt.Printf("\t%s\n", loc)
		}
		fmt.Println()
		fmt.Printf("To add new resources, please submit issue at: %s\n", projectHome)
		os.Exit(1)
	}
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

// extracts 'windows.location' value from <script></script> element tag
// in case of any err, empty string is returned.
func extractEpnRedirect(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return ""
	}
	for _, script := range visit(nil, doc) {
		for _, line := range strings.Split(script, "\n") {
			line = strings.Trim(line, " \t")
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			groups := wlocRe.FindStringSubmatch(line)
			if len(groups) > 1 {
				return groups[1]
			}
		}
	}
	return ""
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

func supportedLocations() []string {
	supported := make([]string, 0, len(locations))
	for k := range locations {
		supported = append(supported, k)
	}
	sort.Strings(supported)
	return supported
}

func isRedirect(status int) bool {
	return status > 299 && status < 400
}
