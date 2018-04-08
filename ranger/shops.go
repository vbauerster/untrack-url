package ranger

import "sort"

func init() {
	registerShops()
}

type CleanUpRule struct {
	Params       []string
	InvertParams bool
	EmptyParams  bool
	EmptyPath    bool
}

var shops = make(map[string]CleanUpRule)

// RegisterShop ...
func RegisterShop(host string, rule CleanUpRule) CleanUpRule {
	prevRule := shops[host]
	shops[host] = rule
	return prevRule
}

// KnownShops ...
func KnownShops() []string {
	list := make([]string, 0, len(shops))
	for k := range shops {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}

func registerShops() {
	// http://ali.pub/2c76pq
	RegisterShop("tmall.aliexpress.com", CleanUpRule{
		Params:       []string{"SearchText"},
		InvertParams: true,
	})
	RegisterShop("ru.aliexpress.com", CleanUpRule{
		Params:       []string{"SearchText"},
		InvertParams: true,
	})
	RegisterShop("www.gearbest.com", CleanUpRule{
		EmptyParams: true,
	})
	RegisterShop("www.coolicool.com", CleanUpRule{
		EmptyParams: true,
	})
	RegisterShop("www.tinydeal.com", CleanUpRule{
		EmptyParams: true,
	})
	RegisterShop("www.banggood.com", CleanUpRule{
		EmptyParams: true,
	})
	RegisterShop("multivarka.pro", CleanUpRule{
		Params:       []string{"q"},
		InvertParams: true,
	})
	// not exactly shop: http://ali.pub/28863g
	RegisterShop("epn.bz", CleanUpRule{
		EmptyParams: true,
	})
	// not exactrly shop: http://ali.pub/1sn27h
	RegisterShop("ali.epn.bz", CleanUpRule{
		EmptyParams: true,
	})
	// not exactrly shop
	RegisterShop("cashback.epn.bz", CleanUpRule{
		EmptyParams: true,
	})
	// not exactrly shop: http://goo.gl/4jTrj4
	RegisterShop("alibonus.com", CleanUpRule{
		EmptyParams: true,
	})
	// not exactrly shop
	RegisterShop("letyshops.ru", CleanUpRule{
		EmptyParams: true,
	})
	// not exactrly shop: https://goo.gl/swMH8e
	RegisterShop("letyshops.com", CleanUpRule{
		EmptyParams: true,
	})
}
