package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ability-sh/abi-micro/grpc"
	"github.com/ability-sh/abi-micro/lrucache"
	"github.com/ability-sh/abi-micro/oss"
	"github.com/ability-sh/abi-micro/redis"

	"github.com/ability-sh/abi-lib/basex"
	"github.com/ability-sh/abi-micro-uri/pb"
	"github.com/spaolacci/murmur3"
	G "google.golang.org/grpc"
)

type URIItem struct {
	URI string `json:"uri"`
	IID int64  `json:"iid"`
}

func (U *URIItem) Key() string {
	m := murmur3.New64()
	m.Write([]byte(U.URI))
	m.Write([]byte(fmt.Sprintf("%d", U.IID)))
	return basex.Base62.Encode(m.Sum(nil))
}

type server struct {
}

func getURIItem(c context.Context, key string) (*URIItem, error) {

	ctx := grpc.GetContext(c)

	//获取服务配置
	URI, err := GetURIService(ctx, SERVICE_URI)

	if err != nil {
		return nil, err
	}

	//加前缀避免微服务间key冲突
	skey := fmt.Sprintf("%s%s", URI.Prefix, key)

	//获取应用内LRU缓存
	cache, err := lrucache.GetCache(ctx, SERVICE_LRUCACHE)

	if err != nil {
		return nil, err
	}

	{
		v, ok := cache.Get(skey)
		if ok {
			return v.(*URIItem), nil
		}
	}

	//获取 REDIS
	rcache, err := redis.GetRedis(ctx, SERVICE_REDIS)

	if err != nil {
		return nil, err
	}

	{
		text, err := rcache.Get(skey)
		if err == nil {
			item := URIItem{}
			err = json.Unmarshal([]byte(text), &item)
			if err == nil {
				cache.Add(key, &item)
				return &item, nil
			}
		}
	}

	//获取 OSS
	oss, err := oss.GetOSS(ctx, SERVICE_OSS)

	if err != nil {
		return nil, err
	}

	{
		b, err := oss.Get(fmt.Sprintf("%s%s", URI.BasePath, key))
		if err == nil {
			item := URIItem{}
			err = json.Unmarshal(b, &item)
			if err == nil {
				cache.Add(key, &item)
				rcache.Set(key, string(b), time.Second*time.Duration(URI.Expires))
				return &item, nil
			}
		}
	}

	return nil, nil
}

func (s *server) Get(c context.Context, task *pb.GetTask) (*pb.GetResult, error) {

	item, err := getURIItem(c, task.Key)

	if err != nil {
		return &pb.GetResult{Errno: ERRNO_INTERNAL_SERVER, Errmsg: err.Error()}, nil
	}

	if item == nil {
		return &pb.GetResult{Errno: ERRNO_NOT_FOUND, Errmsg: "不存在的短链接"}, nil
	}

	return &pb.GetResult{Errno: ERRNO_OK, Uri: item.URI}, nil

}

func (s *server) Set(c context.Context, task *pb.SetTask) (*pb.SetResult, error) {

	item := URIItem{URI: task.Uri, IID: 0}

	key := item.Key()

	ctx := grpc.GetContext(c)

	//获取服务配置
	URI, err := GetURIService(ctx, SERVICE_URI)

	if err != nil {
		return &pb.SetResult{Errno: ERRNO_INTERNAL_SERVER, Errmsg: err.Error()}, nil
	}

	for {

		rs, err := getURIItem(c, key)

		if err != nil {
			return &pb.SetResult{Errno: ERRNO_INTERNAL_SERVER, Errmsg: err.Error()}, nil
		}

		if rs == nil {
			break
		}

		if rs.URI == task.Uri {
			return &pb.SetResult{Errno: ERRNO_OK, Key: key}, nil
		} else {
			item.IID = URI.IID.NewID()
			key = item.Key()
		}
	}

	//获取 OSS
	oss, err := oss.GetOSS(ctx, SERVICE_OSS)

	if err != nil {
		return &pb.SetResult{Errno: ERRNO_INTERNAL_SERVER, Errmsg: err.Error()}, nil
	}

	b, _ := json.Marshal(&item)

	err = oss.Put(fmt.Sprintf("%s%s", URI.BasePath, key), b, nil)

	if err != nil {
		return &pb.SetResult{Errno: ERRNO_INTERNAL_SERVER, Errmsg: err.Error()}, nil
	}

	b, err = oss.Get(fmt.Sprintf("%s%s", URI.BasePath, key))

	if err != nil {
		return &pb.SetResult{Errno: ERRNO_INTERNAL_SERVER, Errmsg: err.Error()}, nil
	}

	skey := fmt.Sprintf("%s%s", URI.Prefix, key)

	//获取应用内LRU缓存
	cache, err := lrucache.GetCache(ctx, SERVICE_LRUCACHE)

	if err == nil {
		cache.Add(skey, &item)
	}

	//获取 REDIS
	rcache, err := redis.GetRedis(ctx, SERVICE_REDIS)

	if err == nil {
		rcache.Set(skey, string(b), time.Second*time.Duration(URI.Expires))
	}

	return &pb.SetResult{Key: key, Errno: ERRNO_OK}, nil
}

func Reg(s *G.Server) {
	pb.RegisterServiceServer(s, &server{})
}
