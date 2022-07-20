package srv

import (
	"fmt"

	"github.com/ability-sh/abi-lib/dynamic"
	"github.com/ability-sh/abi-lib/iid"
	"github.com/ability-sh/abi-micro/micro"
)

type URIService struct {
	config   interface{} `json:"-"`
	name     string      `json:"-"`
	Prefix   string      `json:"prefix"`
	BasePath string      `json:"basePath"`
	Aid      int64       `json:"aid"`     //区域ID
	Nid      int64       `json:"nid"`     //节点ID
	Expires  int64       `json:"expires"` //过期秒数
	IID      *iid.IID    `json:"-"`
}

func newURIService(name string, config interface{}) *URIService {
	return &URIService{name: name, config: config}
}

/**
* 服务名称
**/
func (s *URIService) Name() string {
	return s.name
}

/**
* 服务配置
**/
func (s *URIService) Config() interface{} {
	return s.config
}

/**
* 初始化服务
**/
func (s *URIService) OnInit(ctx micro.Context) error {

	dynamic.SetValue(s, s.config)

	s.IID = iid.NewIID(s.Aid, s.Nid)

	return nil
}

/**
* 校验服务是否可用
**/
func (s *URIService) OnValid(ctx micro.Context) error {
	return nil
}

func (s *URIService) Recycle() {

}

func GetURIService(ctx micro.Context, name string) (*URIService, error) {
	s, err := ctx.GetService(name)
	if err != nil {
		return nil, err
	}
	ss, ok := s.(*URIService)
	if ok {
		return ss, nil
	}
	return nil, fmt.Errorf("service %s not instanceof *URIService", name)
}
