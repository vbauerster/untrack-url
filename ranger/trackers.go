package ranger

import (
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func init() {
	registerTrackers()
}

// https://regex101.com/r/kv1rVs/1
var wlocRe = regexp.MustCompile(`(?:window|document)\.location\s*=\s*['"](.*?)['"]`)
var trackers = make(map[string]ExtractTarget)

// RegisterTracker ...
func RegisterTracker(host string, fn ExtractTarget) ExtractTarget {
	prevFn := trackers[host]
	trackers[host] = fn
	return prevFn
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

func registerTrackers() {
	paramExtractor := func(pkey string) ExtractTarget {
		return func(trackURL *url.URL) (*url.URL, error) {
			pval := trackURL.Query().Get(pkey)
			if targetURL, err := url.Parse(pval); err == nil && targetURL.String() != "" {
				return targetURL, nil
			} else {
				return trackURL, errors.Wrap(
					UntrackErr{errors.Errorf("malformed url at %s=%q", pkey, pval)},
					"paramExtractor",
				)
			}
		}
	}

	//  http://ali.pub/ahgiu
	RegisterTracker("s.click.aliexpress.com", paramExtractor("dl_target_url"))
	// http://fas.st/mKKaRE
	RegisterTracker("ad.admitad.com", paramExtractor("ulp"))
	RegisterTracker("lenkmio.com", paramExtractor("ulp"))
	// http://ali.ski/gkMqy
	RegisterTracker("alitems.com", paramExtractor("ulp"))

	// https://www.youtube.com/redirect?event=video_description&v=p91MiGjZ4wY&q=https%3A%2F%2Fgoo.gl%2FUxgRwh&redir_token=vr-M7CVvynN9nfR23mPw-oeP5HR8MTUyMzI3MTIxNUAxNTIzMTg0ODE1
	RegisterTracker("www.youtube.com", paramExtractor("q"))

	// http://ali.pub/2c753s
	// https://goo.gl/VLb3Xd
	RegisterTracker("epnclick.ru", extractEpnRedirect)
	// http://ali.pub/2c76pq
	RegisterTracker("shopeasy.by", extractEpnRedirect)
}

// extracts 'windows.location' value from <script></script> element tag
func extractEpnRedirect(trackURL *url.URL) (*url.URL, error) {
	resp, err := http.Get(trackURL.String())
	if err != nil {
		return nil, errors.Wrap(UntrackErr{err}, "epn")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(
			UntrackErr{errors.Errorf("expected http status ok, got %q", resp.Status)},
			"epn",
		)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, errors.Wrap(UntrackErr{err}, "epn")
	}
	for _, script := range visit(nil, doc) {
		for _, line := range strings.Split(script, "\n") {
			line = strings.Trim(line, " \t")
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			groups := wlocRe.FindStringSubmatch(line)
			if len(groups) > 1 {
				if targetURL, err := url.Parse(groups[1]); err == nil && targetURL.String() != "" {
					if to := targetURL.Query().Get("to"); to != "" {
						return url.Parse(to)
					}
					return targetURL, nil
				} else {
					return trackURL, errors.Wrap(
						UntrackErr{errors.Errorf("malformed url %q", groups[1])},
						"epn",
					)
				}
			}
		}
	}
	return trackURL, errors.Wrap(
		UntrackErr{errors.New("redirect location not found")},
		"epn",
	)
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
