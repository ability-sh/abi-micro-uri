package srv

import (
	"github.com/ability-sh/abi-micro/micro"
)

func init() {
	micro.Reg("uv-uri", func(name string, config interface{}) (micro.Service, error) {
		return newURIService(name, config), nil
	})
}
