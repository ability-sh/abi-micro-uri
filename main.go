package main

import (
	"log"
	"net"
	"os"

	"github.com/ability-sh/abi-micro-uri/srv"
	"github.com/ability-sh/abi-micro/grpc"
	_ "github.com/ability-sh/abi-micro/logger"
	"github.com/ability-sh/abi-micro/runtime"
	"google.golang.org/grpc/reflection"
)

func main() {

	p, err := runtime.NewFilePayload("./config.yaml", runtime.NewPayload())

	if err != nil {
		log.Panicln(err)
	}

	addr := os.Getenv("ABI_MICRO_ADDR")

	if addr == "" {
		addr = ":8082"
	}

	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Panicln(err)
	}

	s := grpc.NewServer(p)

	srv.Reg(s)

	reflection.Register(s)

	log.Println(addr)

	s.Serve(lis)
}
