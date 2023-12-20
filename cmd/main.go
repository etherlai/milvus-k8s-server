package main

import (
	"fmt"
	"milvus-k8s-server/pkg"
	"milvus-k8s-server/pkg/configs"
	"milvus-k8s-server/pkg/log"
)

func main() {
	log.Init()
	config, err := configs.NewConfig("/milvus/configs")
	if err != nil {
		// run by default, just printing warning.
		fmt.Println("[WARN] load config file failed, running in default setting", err.Error())
	}

	s, err := pkg.NewServer(config)
	if err != nil {
		log.Logger.Fatal("new server failed", err.Error())
	}
	if err := s.Start(); err != nil {
		log.Logger.Fatal("start server failed", err.Error())
	}
}
