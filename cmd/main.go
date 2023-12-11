package main

import (
	"fmt"
	"log"
	"milvus-k8s-server/pkg"
	"milvus-k8s-server/pkg/configs"
)

func main() {
	config, err := configs.NewConfig(".bw_config")
	if err != nil {
		// run by default, just printing warning.
		fmt.Println("[WARN] load config file failed, running in default setting", err.Error())
	}

	s, err := pkg.NewServer(config)
	if err != nil {
		log.Fatal("new server failed", err.Error())
	}
	if err := s.Start(); err != nil {
		log.Fatal("start server failed", err.Error())
	}
}
