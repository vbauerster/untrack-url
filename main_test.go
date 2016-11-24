package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestFollow(t *testing.T) {
	lparam := "dl_target_url"
	ref := "https://ru.aliexpress.com"
	ts := setupTestServer(lparam, ref)
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	debug = true
	redirectHosts[tsURL.Host] = lparam

	gotRef := follow(ts.URL + "/redirect")
	t.Logf("gotRef = %+v\n", gotRef)
	if gotRef != ref {
		t.Errorf("\nExpected ref: %q\nGot ref: %q\n", ref, gotRef)
	}
}

func TestRemoveAds(t *testing.T) {
	ali := "https://ru.aliexpress.com/store/product/Original-Xiaomi-Mi5s-Mi-5S-3GB-RAM-64GB-ROM-Mobile-Phone-Snapdragon-821-5-15-1920x1080/311331_32740701280.html?aff_platform=aaf&sk=VnYZvQVf%3A&cpt=1479995730630&dp=bc1f2bd78b3dbf51260453f6b915ce98&af=288795&cv=47843&afref=&aff_trace_key=e8364f22d3e546fcafe1cf5b61b9519a-1479995730630-04043-VnYZvQVf"
	var tests = []struct {
		in  string
		out string
		dir directive
	}{
		{
			ali,
			"https://ru.aliexpress.com/store/product/Original-Xiaomi-Mi5s-Mi-5S-3GB-RAM-64GB-ROM-Mobile-Phone-Snapdragon-821-5-15-1920x1080/311331_32740701280.html",
			directive{NoQuery: true},
		},
		{
			ali,
			"https://ru.aliexpress.com",
			directive{NoQuery: true, NoPath: true},
		},
		{
			ali,
			"",
			directive{ParamsToDel: []string{"cpt", "aff_trace_key"}},
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
		} else if got != test.out {
			t.Errorf("\nExpected: %q\nGot: %q\n", test.out, got)
		}
	}
}

func setupTestServer(lparam, ref string) *httptest.Server {
	v := url.Values{lparam: {ref}}
	mux := http.NewServeMux()
	mux.Handle("/redirect", http.RedirectHandler("/ref?"+v.Encode(), 302))
	mux.Handle("/ref", http.RedirectHandler(ref, 302))
	return httptest.NewServer(mux)
}
