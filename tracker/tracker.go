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

var Debug bool

var ErrMaxRedirect = fmt.Errorf("tracker: max redirects (%d) followed", maxRedirects)

type ExtractTarget func(tracker *url.URL) (*url.URL, error)

var trackers = make(map[string]ExtractTarget)

type CleanUpRule struct {
	Params       []string
	InvertParams bool
	EmptyParams  bool
	EmptyPath    bool
}

var knownShops = map[string]CleanUpRule{
	// http://ali.pub/2c76pq
	"tmall.aliexpress.com": CleanUpRule{
		Params:       []string{"SearchText"},
		InvertParams: true,
	},
	"ru.aliexpress.com": CleanUpRule{
		Params:       []string{"SearchText"},
		InvertParams: true,
	},
	"www.gearbest.com": CleanUpRule{
		EmptyParams: true,
	},
	"www.coolicool.com": CleanUpRule{
		EmptyParams: true,
	},
	"www.tinydeal.com": CleanUpRule{
		EmptyParams: true,
	},
	"www.banggood.com": CleanUpRule{
		EmptyParams: true,
	},
	"multivarka.pro": CleanUpRule{
		Params:       []string{"q"},
		InvertParams: true,
	},
	// not exactly shop: http://ali.pub/28863g
	"epn.bz": CleanUpRule{
		EmptyParams: true,
	},
	// not exactrly shop: http://ali.pub/1sn27h
	"ali.epn.bz": CleanUpRule{
		EmptyParams: true,
	},
	// not exactrly shop
	"cashback.epn.bz": CleanUpRule{
		EmptyParams: true,
	},
	// not exactrly shop: http://goo.gl/4jTrj4
	"alibonus.com": CleanUpRule{
		EmptyParams: true,
	},
	// not exactrly shop
	"letyshops.ru": CleanUpRule{
		EmptyParams: true,
	},
	// not exactrly shop: https://goo.gl/swMH8e
	"letyshops.com": CleanUpRule{
		EmptyParams: true,
	},
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
		if targetURL, err := checkNestedTrackers(trackURL, nil); err == nil {
			if _, ok := knownShops[targetURL.Host]; ok {
				return targetURL, err
			} else {
				trackURL = targetURL
			}
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

			if Debug {
				fmt.Println("get:", trackURL)
				fmt.Println("loc:", loc)
				fmt.Println()
			}

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

func checkNestedTrackers(trackURL *url.URL, err error) (*url.URL, error) {
	fn, ok := trackers[trackURL.Host]
	if !ok {
		return trackURL, err
	}
	if Debug {
		fmt.Printf("Intercepted tracker: %q\n", trackURL.Host)
	}
	return checkNestedTrackers(fn(trackURL))
}

func Untrack(rawurl string) (string, error) {
	if !strings.Contains(rawurl, "://") && !strings.HasPrefix(rawurl, "//") {
		rawurl = "//" + rawurl
	}
	targetURL, err := follow(rawurl)
	if err != nil {
		return "", err
	}
	if rule, ok := knownShops[targetURL.Host]; ok {
		if Debug {
			fmt.Printf("applying rule %+v to %q\n", rule, targetURL.String())
		}
		if rule.EmptyParams {
			targetURL.RawQuery = ""
		} else if len(rule.Params) != 0 {
			values := targetURL.Query()
			toDelKeys := make(map[string]bool, len(values))
			for k := range values {
				toDelKeys[k] = rule.InvertParams
			}
			for _, k := range rule.Params {
				toDelKeys[k] = !rule.InvertParams
			}
			for k, toDel := range toDelKeys {
				if toDel {
					values.Del(k)
				}
			}
			targetURL.RawQuery = values.Encode()
		}

		if rule.EmptyPath {
			targetURL.Path = ""
		}
	} else if Debug {
		fmt.Printf("host: %q not found in knownShops\n", targetURL.Host)
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
