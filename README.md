# untrack-url [![Build Status](https://travis-ci.org/vbauerster/untrack-url.svg?branch=master)](https://travis-ci.org/vbauerster/untrack-url)

**Why?**

If you follow [such](http://ali.ski/gkMqy) shopping url, and commit a purchase, some advertising party will earn some money on you.

However, you can prevent this by feeding the short url into `untrack-url` tool:

```
$ untrack-url http://ali.ski/gkMqy
```

all tracking query params will be removed and **nobody** will earn money on you, except the seller :innocent:

## Installation
`untrack-url` requires Go 1.7 or later.
```
$ go get -u github.com/vbauerster/untrack-url
```

## Usage
```
Usage: untrack-url [OPTIONS] URL

OPTIONS:
  -d    debug: print debug info, implies -p
  -p    print only: don't open URL in browser
  -v    print version number

Known trackers:

        ad.admitad.com
        alitems.com
        epnclick.ru
        lenkmio.com
        s.click.aliexpress.com
        shopeasy.by
        www.youtube.com

Known shops:

        ali.epn.bz
        alibonus.com
        cashback.epn.bz
        epn.bz
        letyshops.com
        letyshops.ru
        multivarka.pro
        ru.aliexpress.com
        tmall.aliexpress.com
        www.banggood.com
        www.coolicool.com
        www.gearbest.com
        www.tinydeal.com
```

## License

[BSD 3-Clause](https://opensource.org/licenses/BSD-3-Clause)
