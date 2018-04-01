# untrack-url [![Build Status](https://travis-ci.org/vbauerster/untrack-url.svg?branch=master)](https://travis-ci.org/vbauerster/untrack-url)

**untrack-url** will inspect where short url points to, and remove any tracking query parameters,
before opening target url in default web browser.

## Installation
`untrack-url` requires Go 1.7 or later.
```
$ go get -u github.com/vbauerster/untrack-url
```

## Usage
```
Usage: untrack-url [OPTIONS] URL

OPTIONS:
  -p    print only: don't open URL in browser
  -v    print version number

Known trackers:

        ad.admitad.com
        epnclick.ru
        lenkmio.com
        s.click.aliexpress.com
        shopeasy.by
```

## License

[BSD 3-Clause](https://opensource.org/licenses/BSD-3-Clause)
