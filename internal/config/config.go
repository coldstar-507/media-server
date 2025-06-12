package config

import (
	"os"
	"strconv"

	"github.com/coldstar-507/router/router_utils"
)

var Config *router_utils.Config

func LoadConfig() *router_utils.Config {
	conf := &router_utils.Config{}

	ip := os.Getenv("SERVER_IP")
	if len(ip) == 0 {
		panic("undefined SERVER_IP")
	}
	place := os.Getenv("SERVER_PLACE")
	if len(place) == 0 {
		panic("undefined SERVER_PLACE")
	}

	nPlace, err := strconv.Atoi(place)
	if err != nil {
		panic(err)
	}

	conf.SERVER_IP = ip
	conf.SERVER_PLACE = uint16(nPlace)
	conf.SERVER_TYPE = router_utils.MEDIA_ROUTER
	Config = conf
	return conf
}
