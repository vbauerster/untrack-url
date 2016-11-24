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

	if gotRef := follow(ts.URL + "/redirect"); gotRef != ref {
		t.Errorf("\nExpected ref: %q\nGot ref: %q\n", ref, gotRef)
	}
}

func TestRemoveKnownAds(t *testing.T) {
	ali := "https://ru.aliexpress.com/store/product/Original-Xiaomi-Mi5s-Mi-5S-3GB-RAM-64GB-ROM-Mobile-Phone-Snapdragon-821-5-15-1920x1080/311331_32740701280.html?aff_platform=aaf&sk=VnYZvQVf%3A&cpt=1479995730630&dp=bc1f2bd78b3dbf51260453f6b915ce98&af=288795&cv=47843&afref=&aff_trace_key=e8364f22d3e546fcafe1cf5b61b9519a-1479995730630-04043-VnYZvQVf"
	gear := "http://www.gearbest.com/cell-phones/pp_471491.html?wid=21&admitad_uid=75984b1e4bfdbbda9d4238493f856147"
	cool := "http://www.coolicool.com/xiaomi-mi-5s-ultrasonic-fingerprint-3gb-ram-32gb-rom-qualcomm-snapdragon-821-215ghz-quad-core-515-g-44250?utm_source=admitad&utm_medium=referral&admitad_uid=953908711f4f569b3e8acdf0f3ef7ba6"
	tiny := "http://www.tinydeal.com/xiaomi-mi-5s-515-fhd-snapdragon-821-quad-core-android-60-4g-phone-px369k7-p-162019.html?admitad_uid=42926493e6f007492b134a881528cd45&utm_source=admitad&utm_medium=referral&utm_campaign=admitad"
	lety := "https://letyshops.ru/Andronews?admitad_uid=ccdfdeb8f26902ba663c79463dbec762&publisher_id=251289"

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
		// url, err := url.Parse(test.in)
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// got := removeAds(test.in)
		// if len(test.dir.ParamsToDel) != 0 {
		// 	url, err := url.Parse(got)
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}
		// 	v := url.Query()
		// 	for _, p := range test.dir.ParamsToDel {
		// 		if _, ok := v[p]; ok {
		// 			t.Errorf("param %q wasn't removed\n", p)
		// 		}
		// 	}
		// }
		if got := removeAds(test.in); got != test.out {
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
