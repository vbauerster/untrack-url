package ranger

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const (
	maxRedirects = 10
)

var Debug bool

type UntrackErr struct {
	error
}

type ExtractTarget func(*url.URL) (*url.URL, error)

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
		targetURL, err := checkNestedTrackers(trackURL, nil)
		if err != nil {
			return targetURL, errors.WithMessage(err, "follow")
		}
		if _, ok := shops[targetURL.Host]; ok {
			return targetURL, nil
		} else {
			trackURL = targetURL
		}
		resp, err := client.Get(trackURL.String())
		if err != nil {
			return nil, errors.Wrap(UntrackErr{err}, "follow")
		}
		resp.Body.Close()

		if isRedirect(resp.StatusCode) {
			locURL, err := resp.Location()
			if err != nil {
				return nil, errors.Wrap(UntrackErr{err}, "follow")
			}

			if Debug {
				fmt.Println("get:", trackURL)
				fmt.Println("loc:", locURL)
				fmt.Println()
			}

			redirectsFollowed++
			if redirectsFollowed > maxRedirects {
				return nil, errors.Wrap(
					UntrackErr{errors.Errorf("max redirects (%d) followed", maxRedirects)},
					"follow",
				)
			}
			trackURL = locURL
		} else {
			return trackURL, nil
		}
	}
}

func checkNestedTrackers(trackURL *url.URL, err error) (*url.URL, error) {
	fn, ok := trackers[trackURL.Host]
	if !ok || err != nil {
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
		return "", errors.WithMessage(err, "Untrack")
	}

	if rule, ok := shops[targetURL.Host]; ok {
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

		return targetURL.String(), nil
	}

	return targetURL.String(), errors.Wrap(
		UntrackErr{errors.Errorf("%q not found in known shops", targetURL.Host)},
		"Untrack",
	)
}

func isRedirect(status int) bool {
	return status > 299 && status < 400
}
