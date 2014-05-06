package module

import (
	"log"
	"time"

	"github.com/hugb/beegecluster/cluster"
	"github.com/hugb/beegecluster/config"
	"github.com/hugb/beegecluster/proxy"
	"github.com/hugb/beegecluster/utils"
)

// Dcoker模块
func StartControllerrModule(joinAddress, serviceAddress, clusterAddress string) {
	// 设置日志格式
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	// 参数检查
	if serviceAddress == "" {
		log.Fatal("Service address is required")
	}
	if clusterAddress == "" {
		log.Fatal("Cluster address is required")
	}
	// 保存配置
	config.JoinAddress = joinAddress
	config.ServiceAddress = serviceAddress
	config.ClusterAddress = clusterAddress
	config.Role = config.ControllerRoleName
	config.Controllers[clusterAddress] = time.Now().Unix()

	// 集群内部通信服务器
	go cluster.NewClusterServer()
	// 注册内部通信命令处理函数
	cluster.ClusterHandlers()
	// 与docker连接断开后处理
	cluster.ClusterSwitcher.Register("disconnect", DockerDisconnection)

	if config.JoinAddress != "" {
		config.Controllers[joinAddress] = time.Now().Unix()
		// 从接入点获取集群的结构
		go cluster.ControllerJoinCluster()
	}

	// 启动代理服务器
	proxy.NewProxyServer()
}

// 我的小弟死了
func DockerDisconnection(c *utils.Connection, data []byte) {
	log.Println("docker:", string(data), "is offline.")
	// 从生死簿中将他的名字抹去
	delete(config.Dockers, string(data))
	log.Println("dockers:", config.Dockers)
}
