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
	if gotRef != ref {
		t.Errorf("Got ref: %q\nExpected ref: %q\n", gotRef, ref)
	}
}

func setupTestServer(lparam, loc string) *httptest.Server {
	v := url.Values{lparam: {loc}}
	mux := http.NewServeMux()
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/deep_link.htm?"+v.Encode(), 302)
	})
	return httptest.NewServer(mux)
}
