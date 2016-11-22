package main

import (
	"log"
	"net/url"
	"strings"
)

type directive struct {
	ParamsToRemove []string
	NoParams       bool
	HostOnly       bool
}

// var knownMarkets = [...]directive{
// 	{"s.click.aliexpress.com", "dl_target_url"},
// 	{"ad.admitad.com", "ulp"},
// }

var redirectHosts = map[string]string{
	"s.click.aliexpress.com": "dl_target_url",
	"ad.admitad.com":         "ulp",
}

var locations = make(map[string]directive)

func init() {
	locations["ru.aliexpress.com"] = directive{NoParams: true}
	locations["www.gearbest.com"] = directive{ParamsToRemove: []string{"wid", "subid"}}
	locations["www.coolicool.com"] = directive{NoParams: true}
	locations["www.tinydeal.com"] = directive{NoParams: true}
	locations["letyshops.ru"] = directive{HostOnly: true}
}

func main() {

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
