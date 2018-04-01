package tracker

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const (
	maxRedirects = 10
)

var ErrMaxRedirect = fmt.Errorf("tracker: max redirects (%d) followed", maxRedirects)

type ExtractTarget func(tracker *url.URL) (*url.URL, error)

var trackers = make(map[string]ExtractTarget)

type directive struct {
	ParamsToDel []string
	NoQuery     bool
	NoPath      bool
	Scheme      string
}

var knownShops = map[string]directive{
	"ru.aliexpress.com": directive{NoQuery: true},
	"www.gearbest.com":  directive{NoQuery: true},
	"www.coolicool.com": directive{NoQuery: true},
	"www.tinydeal.com":  directive{NoQuery: true},
	"www.banggood.com":  directive{NoQuery: true},
	"letyshops.ru":      directive{NoQuery: true, NoPath: true, Scheme: "https"},
	"cashback.epn.bz":   directive{NoQuery: true, NoPath: true},
	"alibonus.com":      directive{NoQuery: true, NoPath: true},
}

// RegisterTracker ...
func RegisterTracker(host string, fn ExtractTarget) ExtractTarget {
	prevFn := trackers[host]
	trackers[host] = fn
	return prevFn
}

func follow(rawurl string) (*url.URL, error) {
	trackURL, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	// number of redirects followed
	var redirectsFollowed int
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for {
		if f, ok := trackers[trackURL.Host]; ok {
			return f(trackURL)
		}
		resp, err := client.Get(trackURL.String())
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		if isRedirect(resp.StatusCode) {
			loc, err := resp.Location()
			if err != nil {
				return nil, err
			}

			// fmt.Println("get:", trackURL)
			// fmt.Println("loc:", loc)
			// fmt.Println()

			redirectsFollowed++
			if redirectsFollowed > maxRedirects {
				return nil, ErrMaxRedirect
			}
			trackURL = loc
		} else {
			return trackURL, nil
		}
	}
}

func Untrack(rawurl string) (string, error) {
	if !strings.Contains(rawurl, "://") && !strings.HasPrefix(rawurl, "//") {
		rawurl = "//" + rawurl
	}
	targetURL, err := follow(rawurl)
	if err != nil {
		return "", err
	}
	if dir, ok := knownShops[targetURL.Host]; ok {
		// fmt.Printf("%s = %+v\n", targetURL.Host, dir)
		if dir.NoQuery {
			targetURL.RawQuery = ""
		} else if len(dir.ParamsToDel) != 0 {
			v := targetURL.Query()
			for _, param := range dir.ParamsToDel {
				v.Del(param)
			}
			targetURL.RawQuery = v.Encode()
		}
		if dir.NoPath {
			targetURL.Path = ""
		}
		if dir.Scheme != "" {
			targetURL.Scheme = dir.Scheme
		}
	}
	return targetURL.String(), nil
}

// KnownTrackers ...
func KnownTrackers() []string {
	list := make([]string, 0, len(trackers))
	for k := range trackers {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}

func isRedirect(status int) bool {
	return status > 299 && status < 400
}
