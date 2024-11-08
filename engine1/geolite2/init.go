// Package geolite2
// @Title
// @Description  geolite2
// @Author sly 2024/9/3
// @Created sly 2024/9/3
package geolite2

import (
	"errors"
	"github.com/oschwald/geoip2-golang"
)

var GeoLiteReader *geoip2.Reader

func InitGeoLite2IP(dbPath string) error {
	geoLite, err := geoip2.Open(dbPath)
	if err != nil {
		return err
	}
	if geoLite == nil {
		return errors.New("InitGeoLite2IP geoLite is nil")
	}
	GeoLiteReader = geoLite
	return nil
}
