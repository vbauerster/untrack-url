package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/vbauerster/untrack-url/ranger"
)

func setupRedirectServer(source, destination string) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle(source, http.RedirectHandler(destination, 302))
	return httptest.NewServer(mux)
}

func setupTestServer(path string, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(path, handler)
	return httptest.NewServer(mux)
}

type testCase struct {
	trackerHost string
	targetKey   string
	dirtyTarget string
	cleanTarget string
	script      string
	handleMaker func(string) func(http.ResponseWriter, *http.Request)
}

func TestUntrack(t *testing.T) {

	troubleMaker := func(string) func(http.ResponseWriter, *http.Request) {
		return func(http.ResponseWriter, *http.Request) {
			t.Fail()
		}
	}

	makeScriptHandler := func(script string) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, _ *http.Request) {
			body := `<!DOCTYPE html>
					<html>
						<head>
							<title>Redirecting...</title>
							<meta charset="utf-8">
						</head>
						<body>
							<script>%s</script>
						</body>
					</html>`
			io.WriteString(w, fmt.Sprintf(body, script))
		}
	}

	cases := map[string]testCase{
		"targetKey: dl_target_url": {
			trackerHost: "s.click.aliexpress.com",
			targetKey:   "dl_target_url",
			dirtyTarget: "https://ru.aliexpress.com/store?a=A&SearchText=phone&b=B&c=C",
			cleanTarget: "https://ru.aliexpress.com/store?SearchText=phone",
			handleMaker: troubleMaker,
		},
		"targetKey: ulp": {
			trackerHost: "ad.admitad.com",
			targetKey:   "ulp",
			dirtyTarget: "https://ru.aliexpress.com/store?a=A&SearchText=phone&b=B&c=C",
			cleanTarget: "https://ru.aliexpress.com/store?SearchText=phone",
			handleMaker: troubleMaker,
		},
		"targetKey: q": {
			trackerHost: "www.youtube.com",
			targetKey:   "q",
			dirtyTarget: "https://ad.admitad.com?ulp=https%3A%2F%2Fru.aliexpress.com%2Fstore%3Fa%3DA%26SearchText%3Dphone%26b%3DB%26c%3DC",
			cleanTarget: "https://ru.aliexpress.com/store?SearchText=phone",
			handleMaker: troubleMaker,
		},
		"epn": {
			trackerHost: "epnclick.ru",
			cleanTarget: "http://www.gearbest.com/cell-phones/pp_470619.html",
			script:      "window.location = 'http://www.gearbest.com/cell-phones/pp_470619.html?wid=21&utm_source=epn';",
			handleMaker: makeScriptHandler,
		},
		"enp with to": {
			trackerHost: "shopeasy.by",
			cleanTarget: "https://tmall.aliexpress.com/w/wholesale-multicooker.html?SearchText=multicooker",
			script:      "document.location='/redirect/cpa/o/p5brt6my0anysg50o8syzaw1yyu1mhxv/?to=https%3A%2F%2Ftmall.aliexpress.com%2Fw%2Fwholesale-multicooker.html%3Fspm%3Da2g02.9334986.kitchen-appliances.8.21154eaexojb3q%26site%3Drus%26SearchText%3Dmulticooker%26needQuery%3Dn%26initiative_id%3DSB_20171210225006%26isCompetitiveProducts%3Dy%26g%3Dy';",
			handleMaker: makeScriptHandler,
		},
	}

	ranger.Debug = true
	for name, tc := range cases {
		setupAndTest(t, name, tc)
	}
}

// setupAndTest ...
func setupAndTest(t *testing.T, name string, tc testCase) {
	ts := setupTestServer("/", tc.handleMaker(tc.script))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fail()
	}
	if tc.targetKey != "" {
		v := url.Values{}
		v.Set(tc.targetKey, tc.dirtyTarget)
		tsURL.RawQuery = v.Encode()
	}
	rs := setupRedirectServer("/", tsURL.String())
	defer rs.Close()

	fn := ranger.RegisterTracker(tc.trackerHost, nil)
	if fn == nil {
		t.Fail()
	}
	ranger.RegisterTracker(tsURL.Host, fn)
	ranger.RegisterTracker(tc.trackerHost, fn)

	if target, err := ranger.Untrack(rs.URL); err == nil {
		if target != tc.cleanTarget {
			t.Errorf("%s: expected: %q, got: %q\n", name, tc.cleanTarget, target)
		}
	} else {
		t.Fail()
	}
}
