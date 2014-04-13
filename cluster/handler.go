package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hugb/beegecontroller/config"
	"github.com/hugb/beegecontroller/utils"
)

func ClusterHandlers() {
	m := map[string]HandlerFunc{
		"docker_greetings":        dockerGreetings,
		"docker_join_cluster":     dockerJoinCluster,
		"controller_join_cluster": controllerJoinCluster,
		"controller_offline":      controllerOffline,
	}
	for cmd, fct := range m {
		if err := ClusterSwitcher.Register(cmd, fct); err != nil {
			log.Printf("register cluster handler[%s] failure:%s.\n", cmd, err)
		} else {
			log.Printf("register cluster hander[%s] success.\n", cmd)
		}
	}
}

// 来自docker问候
func dockerGreetings(c *utils.Connection, data []byte) {
	c.Src = string(data)
	config.CS.ClusterServer.Docker[string(data)] = time.Now().Unix()
	log.Println("docker:", string(data), "is online.")
	log.Println("dockers:", config.CS.ClusterServer.Docker)
}

func dockerJoinCluster(c *utils.Connection, data []byte) {
	// 向集群结构配置里面添加新成员
	config.CS.ClusterServer.Docker[string(data)] = time.Now().Unix()
	// 返回集群中的controller给新成员，以便新成员连接到其他controller
	b, err := json.Marshal(config.CS.ClusterServer.Controller)
	if err != nil {
		c.SendFailsResult("docker_join_cluster", fmt.Sprintf("%s", err))
	} else {
		c.SendSuccessResultBytes("docker_join_cluster", b)
	}
}

func controllerJoinCluster(c *utils.Connection, data []byte) {
	address := string(data)
	// 向集群结构配置里面添加新成员
	config.CS.ClusterServer.Controller[address] = time.Now().Unix()
	// 返回集群中的controller给新成员，以便新成员挨个通知其加入了集群
	b, err := json.Marshal(config.CS.ClusterServer.Controller)
	if err != nil {
		c.SendFailsResult("controller_join_cluster", fmt.Sprintf("%s", err))
	} else {
		c.SendSuccessResultBytes("controller_join_cluster", b)
	}
	// 向连接到自己的所有docker广播有controller加入集群
	// todo:连接到controller的新成员也会收到广播，应避免和本次返回的结果混淆
	time.Sleep(1 * time.Second)
	// 1秒后本连接已经关闭，广播和本次结果不会混淆
	ClusterSwitcher.broadcast <- []byte(fmt.Sprintf("%s %s", address, "controller_join_cluster"))
	log.Println("controllers", config.CS.ClusterServer.Controller)
}

// controller离线
func controllerOffline(c *utils.Connection, data []byte) {
	delete(config.CS.ClusterServer.Controller, string(data))
	log.Println("controller:", string(data), "is offline.")
}

func DockerDisconnection(c *utils.Connection, data []byte) {
	log.Println("docker:", string(data), "is offline.")
	delete(config.CS.ClusterServer.Docker, string(data))
	log.Println("dockers:", config.CS.ClusterServer.Docker)
}

func ControllerDisconnection(c *utils.Connection, data []byte) {
	log.Println("controller:", string(data), "is offline.")
	ClusterSwitcher.Broadcast(utils.PacketByes(append(data, " controller_offline"...)))
}
