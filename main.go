package main

import (
	"flag"
	"log"

	"github.com/hugb/beegecontroller/cluster"
	"github.com/hugb/beegecontroller/proxy"
	//"github.com/hugb/beegecontroller/sync"
)

func main() {
	var (
		accessPoint    = flag.String("a", "", "Access Point")
		proxyAddress   = flag.String("p", "", "Proxy Address")
		clusterAddress = flag.String("c", "", "Cluster Address")
	)
	flag.Parse()

	if *accessPoint != "" {
		// 从接入点获取集群的结构
	}

	if *clusterAddress != "" {
		go cluster.NewClusterServer(*clusterAddress)
	} else {
		// 退出
		log.Fatal("Cluster address can not be empty.")
	}

	if *proxyAddress != "" {
		// 启动代理服务器
		proxy.NewProxyServer(*proxyAddress)
	} else {
		// 退出
		log.Fatal("Proxy address can not be empty.")
	}
}
