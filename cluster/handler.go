package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dotcloud/docker/engine"

	"github.com/hugb/beegecontroller/config"
	"github.com/hugb/beegecontroller/utils"
)

func ClusterHandlers() {
	m := map[string]HandlerFunc{
		"heartbeat":               heartbeat,
		"docker_status":           dockerStatus,
		"docker_event":            dockerEvent,
		"docker_images":           dockerImages,
		"docker_containers":       dockerContainers,
		"docker_greetings":        dockerGreetings,
		"docker_join_cluster":     dockerJoinCluster,
		"controller_join_cluster": controllerJoinCluster,
		"controller_offline":      controllerOffline,
	}
	for cmd, fct := range m {
		if err := ClusterSwitcher.Register(cmd, fct); err != nil {
			log.Printf("Register cluster handler[%s] failure:%s", cmd, err)
		} else {
			log.Printf("Register cluster hander[%s] success", cmd)
		}
	}
}

// 心跳
func heartbeat(c *utils.Connection, data []byte) {
	c.SendSuccessResultString("heartbeat", "")
}

// docker主机状态
func dockerStatus(c *utils.Connection, data []byte) {
	log.Println("Status:", string(data))
}

// docker事件
func dockerEvent(c *utils.Connection, data []byte) {
	log.Println("Event:", string(data))
}

// docker主机上的镜像
func dockerImages(c *utils.Connection, data []byte) {
	dst := engine.NewTable("", 0)
	if _, err := dst.ReadListFrom(data); err != nil {
		log.Println("Read table error:", err)
	}
	if content, err := dst.ToListString(); err != nil {
		log.Println("Table to string error:", err)
	} else {
		log.Println("Images:", content)
	}
}

// docker主机上的容器
func dockerContainers(c *utils.Connection, data []byte) {
	dst := engine.NewTable("", 0)
	if _, err := dst.ReadListFrom(data); err != nil {
		log.Println("Read table error:", err)
	}
	if content, err := dst.ToListString(); err != nil {
		log.Println("Table to string error:", err)
	} else {
		log.Println("Containers:", content)
	}
}

// 注册资料
func dockerGreetings(c *utils.Connection, data []byte) {
	c.Src = string(data)
	config.CS.ClusterServer.Docker[string(data)] = time.Now().Unix()
	log.Println("Docker:", string(data), "is online.")
	log.Println("Dockers:", config.CS.ClusterServer.Docker)
}

// 我收了个小弟
// todo:是否要审核
// 由于其知道我的名字我默认相信他
func dockerJoinCluster(c *utils.Connection, data []byte) {
	// 向集群结构配置里面添加新成员
	config.CS.ClusterServer.Docker[string(data)] = time.Now().Unix()
	// 返回组织中领导层所有人姓名以便小弟有事时着他们
	b, err := json.Marshal(config.CS.ClusterServer.Controller)
	if err != nil {
		// 我收集的资料有误
		c.SendFailsResult("docker_join_cluster", fmt.Sprintf("%s", err))
	} else {
		// 告知其组织的领导层所有人员姓名
		c.SendSuccessResultBytes("docker_join_cluster", b)
	}
}

// 结拜了个兄弟
func controllerJoinCluster(c *utils.Connection, data []byte) {
	address := string(data)
	// 把他名字记下来
	config.CS.ClusterServer.Controller[address] = time.Now().Unix()
	// 把我以前结拜的所有兄弟告诉他，让他们也认识一下
	b, err := json.Marshal(config.CS.ClusterServer.Controller)
	if err != nil {
		c.SendFailsResult("controller_join_cluster", fmt.Sprintf("%s", err))
	} else {
		c.SendSuccessResultBytes("controller_join_cluster", b)
	}
	// 先断开连接，以免其收到广播我发给小弟的通知
	c.Conn.Close()
	// 告知我的所有小弟，我认了个兄弟，以后的进贡也要给他们一份
	ClusterSwitcher.broadcast <- []byte(fmt.Sprintf("%s %s", address, "controller_join_cluster"))
	log.Println("Controllers", config.CS.ClusterServer.Controller)
}

// 小弟说我结拜的兄弟死了
func controllerOffline(c *utils.Connection, data []byte) {
	log.Println("controller:", string(data), "is offline.")
	// 从生死簿中将他的名字抹去
	delete(config.CS.ClusterServer.Controller, string(data))
	log.Println("Controllers:", config.CS.ClusterServer.Controller)
}

// 我的小弟死了
func DockerDisconnection(c *utils.Connection, data []byte) {
	log.Println("docker:", string(data), "is offline.")
	// 从生死簿中将他的名字抹去
	delete(config.CS.ClusterServer.Docker, string(data))
	log.Println("dockers:", config.CS.ClusterServer.Docker)
}

// controller连接到docker的连接断开，广播给其他controller
func ControllerDisconnection(c *utils.Connection, data []byte) {
	log.Println("controller:", string(data), "is offline.")
	ClusterSwitcher.Broadcast(utils.PacketByes(append(data, " controller_offline"...)))
}
