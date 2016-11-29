package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func setupTestServer(ref, param string) *httptest.Server {
	v := url.Values{param: {ref}}
	mux := http.NewServeMux()
	mux.Handle("/redirect", http.RedirectHandler("/ref?"+v.Encode(), 302))
	mux.Handle("/ref", http.RedirectHandler(ref, 302))
	return httptest.NewServer(mux)
}

func setupTestEpnServer(content string) *httptest.Server {
	body := `<!DOCTYPE html>
	<html>
	<head>
		<title>Redirecting...</title>
			<meta charset="utf-8">
			</head>
			<body>
					<script>content</script>
			</body>
	</html>`
	mux := http.NewServeMux()
	mux.Handle("/redirect", http.RedirectHandler("/ref", 302))
	mux.HandleFunc("/ref", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, strings.Replace(body, "content", content, -1))
	})
	return httptest.NewServer(mux)
}

func TestFollow(t *testing.T) {
	params := make([]string, 0, len(redirectHosts))
	for _, v := range redirectHosts {
		params = append(params, v)
	}

	dummyRef := "http://dummy.org"
	for _, param := range params {
		ts := setupTestServer(dummyRef, param)
		defer ts.Close()

		tsURL, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatal(err)
		}

		redirectHosts[tsURL.Host] = param

		if ref := follow(ts.URL + "/redirect"); ref != dummyRef {
			t.Errorf("\nExpected ref: %q\nGot ref: %q\n", dummyRef, ref)
		}
	}
}

func TestRemoveAds(t *testing.T) {
	ali := "https://ru.aliexpress.com/store/product/Original-Xiaomi-Mi5s-Mi-5S-3GB-RAM-64GB-ROM-Mobile-Phone-Snapdragon-821-5-15-1920x1080/311331_32740701280.html?aff_platform=aaf&sk=VnYZvQVf%3A&cpt=1479995730630&dp=bc1f2bd78b3dbf51260453f6b915ce98&af=288795&cv=47843&afref=&aff_trace_key=e8364f22d3e546fcafe1cf5b61b9519a-1479995730630-04043-VnYZvQVf"
	gear := "http://www.gearbest.com/cell-phones/pp_471491.html?wid=21&admitad_uid=75984b1e4bfdbbda9d4238493f856147"
	cool := "http://www.coolicool.com/xiaomi-mi-5s-ultrasonic-fingerprint-3gb-ram-32gb-rom-qualcomm-snapdragon-821-215ghz-quad-core-515-g-44250?utm_source=admitad&utm_medium=referral&admitad_uid=953908711f4f569b3e8acdf0f3ef7ba6"
	tiny := "http://www.tinydeal.com/xiaomi-mi-5s-515-fhd-snapdragon-821-quad-core-android-60-4g-phone-px369k7-p-162019.html?admitad_uid=42926493e6f007492b134a881528cd45&utm_source=admitad&utm_medium=referral&utm_campaign=admitad"
	lety := "http://letyshops.ru/Andronews?admitad_uid=ccdfdeb8f26902ba663c79463dbec762&publisher_id=251289"

	var tests = []struct {
		in  string
		out string
	}{
		{
			ali,
			"https://ru.aliexpress.com/store/product/Original-Xiaomi-Mi5s-Mi-5S-3GB-RAM-64GB-ROM-Mobile-Phone-Snapdragon-821-5-15-1920x1080/311331_32740701280.html",
		},
		{
			gear,
			"http://www.gearbest.com/cell-phones/pp_471491.html",
		},
		{
			cool,
			"http://www.coolicool.com/xiaomi-mi-5s-ultrasonic-fingerprint-3gb-ram-32gb-rom-qualcomm-snapdragon-821-215ghz-quad-core-515-g-44250",
		},
		{
			tiny,
			"http://www.tinydeal.com/xiaomi-mi-5s-515-fhd-snapdragon-821-quad-core-android-60-4g-phone-px369k7-p-162019.html",
		},
		{
			lety,
			"https://letyshops.ru",
		},
	}

	for _, test := range tests {
		if got := removeAds(test.in); got != test.out {
			t.Errorf("\nExpected: %q\nGot: %q\n", test.out, got)
		}
	}
}

func TestRemoveAdsArbitraryParam(t *testing.T) {
	var tests = []struct {
		in  string
		dir directive
	}{
		{
			"https://example.org/test?spy_ad=toberemoved",
			directive{ParamsToDel: []string{"spy_ad"}},
		},
		{
			"https://example.org/test?spy_ad=toberemoved&wid=foo",
			directive{ParamsToDel: []string{"wid"}},
		},
	}
	for _, test := range tests {
		url, err := url.Parse(test.in)
		if err != nil {
			t.Fatal(err)
		}
		locations[url.Host] = test.dir
		got := removeAds(test.in)
		if len(test.dir.ParamsToDel) != 0 {
			url, err := url.Parse(got)
			if err != nil {
				t.Fatal(err)
			}
			v := url.Query()
			for _, p := range test.dir.ParamsToDel {
				if _, ok := v[p]; ok {
					t.Errorf("param %q wasn't removed\n", p)
				}
			}
		}
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://golang.org", "https://golang.org"},
		{"https://golang.org:443/test", "https://golang.org:443/test"},
		{"localhost:8080/test", "https://localhost:8080/test"},
		{"localhost:80/test", "http://localhost:80/test"},
		{"//localhost:8080/test", "https://localhost:8080/test"},
		{"//localhost:80/test", "http://localhost:80/test"},
	}

	for _, test := range tests {
		u := parseURL(test.in)
		if u.String() != test.want {
			t.Errorf("Given: %s\nwant: %s\ngot: %s", test.in, test.want, u.String())
		}
	}
}

func TestExtractEpnRedirect(t *testing.T) {
	debug = true
	tests := []struct {
		content string
		want    string
	}{
		{
			"window.location = 'http://www.gearbest.com/cell-phones/pp_470619.html?wid=21&utm_source=epn'",
			"http://www.gearbest.com/cell-phones/pp_470619.html?wid=21&utm_source=epn",
		},
		{
			"\n\twindow.location='http://www.gearbest.com/cell-phones/pp_470619.html?wid=21&utm_source=epn';\n",
			"http://www.gearbest.com/cell-phones/pp_470619.html?wid=21&utm_source=epn",
		},
		{
			`window.location="http://www.gearbest.com/cell-phones/pp_470619.html?wid=21&utm_source=epn";`,
			"http://www.gearbest.com/cell-phones/pp_470619.html?wid=21&utm_source=epn",
		},
	}

	for _, test := range tests {
		ts := setupTestEpnServer(test.content)
		defer ts.Close()

		url := extractEpnRedirect(ts.URL + "/redirect")
		if url != test.want {
			t.Errorf("Expected loc: %q\nGot loc: %q\n", test.want, url)
		}
	}
}
