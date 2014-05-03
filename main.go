///////////////////////////////////////////////////////////////////
/*                          Controller                           */
///////////////////////////////////////////////////////////////////
// 提供API路由选择，以及集群管理
package main

import (
	"flag"
	"log"
	"time"

	"github.com/hugb/beegecluster/cluster"
	"github.com/hugb/beegecluster/config"
	"github.com/hugb/beegecluster/proxy"
)

func main() {
	var (
		joinPoint      = flag.String("j", "", "Join Point Address")
		serviceAddress = flag.String("p", "", "Proxy Address")
		clusterAddress = flag.String("c", "", "Cluster Address")
	)
	flag.Parse()

	if *serviceAddress == "" {
		log.Fatal("Proxy address is required.")
	}
	if *clusterAddress == "" {
		log.Fatal("Cluster address is required.")
	}

	// 保存配置
	config.Role = config.ControllerRoleName

	config.CS.JoinPoint = *joinPoint
	config.CS.ServiceAddress = *serviceAddress
	config.CS.ClusterAddress = *clusterAddress
	config.CS.ClusterServer.Controller[*clusterAddress] = time.Now().Unix()

	if config.CS.JoinPoint != "" {
		config.CS.ClusterServer.Controller[*joinPoint] = time.Now().Unix()
		// 从接入点获取集群的结构
		go cluster.ControllerJoinCluster()
	}

	// 集群内部通信服务器
	go cluster.NewClusterServer()
	// 注册内部通信命令处理函数
	cluster.ClusterHandlers()
	// 与docker连接断开后处理
	cluster.ClusterSwitcher.Register("disconnect", cluster.DockerDisconnection)
	// 启动代理服务器
	proxy.NewProxyServer()
}

////////////////////////////////////////////////////////////////////////////
/*                           作为docker一部分                               */
////////////////////////////////////////////////////////////////////////////
/*
package main

import (
	"flag"
	"log"
	"time"

	"github.com/hugb/beegecluster/cluster"
	"github.com/hugb/beegecluster/config"
)

func main() {

	var (
		joinPoint      = flag.String("j", "", "Join Point")
		serviceAddress = flag.String("s", "", "Listen address")
	)
	flag.Parse()

	if *joinPoint == "" {
		log.Fatal("join point is required.")
	}
	if *serviceAddress == "" {
		log.Fatal("serivce address is required.")
	}

	// 保存配置
	config.CS.JoinPoint = *joinPoint
	config.CS.ServiceAddress = *serviceAddress
	config.CS.ClusterServer.Controller[*joinPoint] = time.Now().Unix()
	config.CS.ClusterServer.Docker[*serviceAddress] = time.Now().Unix()

	// 与controller连接断开后，将向连接的所有controller广播
	cluster.ClusterSwitcher.Register("disconnect", cluster.ControllerDisconnection)

	// docker加入集群
	cluster.DockerJoinCluster()
}
*/
/*

///////////////////////////////////////////////////////////////////////
flJoinPoint = flag.String([]string{"j", "-joinpoint"}, "", "Cluster join  point")
// cluster module
if *flJoinPoint == "" {
	log.Fatal("Cluster join point is required.")
}

// 保存配置
config.CS.JoinPoint = *flJoinPoint
config.CS.ClusterAddress = flHosts.GetAll()[0]
config.CS.ClusterServer.Controller[*flJoinPoint] = time.Now().Unix()
config.CS.ClusterServer.Docker[flHosts.GetAll()[0]] = time.Now().Unix()

// 与controller连接断开后，将向连接的所有controller广播
cluster.ClusterSwitcher.Register("disconnect", cluster.ControllerDisconnection)

// docker加入集群
cluster.DockerJoinCluster(eng)
//////////////////////////////////////////////////////////////////////
*/
