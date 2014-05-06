package module

import (
	"log"
	"strings"
	"time"

	"github.com/dotcloud/docker/engine"

	"github.com/hugb/beegecluster/cluster"
	"github.com/hugb/beegecluster/config"
	"github.com/hugb/beegecluster/utils"
)

// Dcoker插件
func StartDockerModule(protoAddrs []string, joinAddress string, eng *engine.Engine) {
	var (
		clusterAddress string
	)
	for _, protoAddr := range protoAddrs {
		protoAddrParts := strings.SplitN(protoAddr, "://", 2)
		if len(protoAddrParts) == 2 && protoAddrParts[0] == "tcp" {
			clusterAddress = protoAddrParts[1]
		}
	}

	if joinAddress == "" {
		log.Fatal("Join address is required")
	}
	if clusterAddress == "" {
		log.Fatal("Cluster address is required")
	}

	// 设置日志格式
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)

	// 保存配置
	config.JoinAddress = joinAddress
	config.Role = config.DockerRoleName
	config.ClusterAddress = clusterAddress
	config.Dockers[clusterAddress] = time.Now().Unix()
	config.Controllers[joinAddress] = time.Now().Unix()

	// 注册内部通信命令处理函数
	cluster.ClusterHandlers()
	// 与controller连接断开后，将向连接的所有controller广播
	cluster.ClusterSwitcher.Register("disconnect", ControllerDisconnection)

	// docker加入集群
	cluster.DockerJoinCluster(eng)
}

// controller连接到docker的连接断开，广播给其他controller
func ControllerDisconnection(c *utils.Connection, data []byte) {
	address := string(data)
	log.Println("controller:", address, "is offline.")
	delete(config.Controllers, address)
	cluster.ClusterSwitcher.Broadcast(utils.PacketByes(append(data, " controller_offline"...)))
}
