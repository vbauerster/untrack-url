package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestFollow(t *testing.T) {
	lparam := "dl_target_url"
	ts := setupTestServer(lparam, "https://ru.aliexpress.com")
	defer ts.Close()

	fmt.Printf("ts.URL = %+v\n", ts.URL)
	fmt.Printf("ts.Config = %+v\n", ts.Config)
	fmt.Printf("ts.Config.Handler = %+v\n", ts.Config.Handler)

	debug = true
	redirectHosts["127.0.0.1"] = lparam
	fmt.Printf("redirectHosts = %+v\n", redirectHosts)

	url, param := follow(parseURL(ts.URL + "/redirect"))
	fmt.Printf("url = %+v\n", url)
	if param != "dl_target_url" {
		t.Errorf("unexpected param: %s", param)
	}
}

func setupTestServer(lparam, loc string) *httptest.Server {
	v := url.Values{lparam: {loc}}
	mux := http.NewServeMux()
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/deep_link.htm?"+v.Encode(), 302)
	})
	ts := httptest.NewServer(mux)
	return ts
}
