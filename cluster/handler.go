package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dotcloud/docker/engine"

	"github.com/hugb/beegecluster/config"
	"github.com/hugb/beegecluster/registry"
	"github.com/hugb/beegecluster/resource"
	"github.com/hugb/beegecluster/utils"
)

func ClusterHandlers() {
	m := map[string]HandlerFunc{
		"heartbeat":                 heartbeat,
		"docker_status":             dockerStatus,
		"docker_event":              dockerEvent,
		"docker_images":             dockerImages,
		"docker_containers":         dockerContainers,
		"docker_greetings":          dockerGreetings,
		"docker_greetings_reply":    dockerGreetingsReply,
		"docker_join":               dockerJoin,
		"controller_join":           controllerJoin,
		"controller_offline":        controllerOffline,
		"controller_join_to_docker": controllerJoinToDocker,
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
	c.SendCommandString("heartbeat", config.ClusterAddress)
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
	for _, env := range dst.Data {
		created, err := strconv.ParseInt(env.Get("Created"), 10, 64)
		if err == nil {
			image := &resource.Image{Host: c.Src, Created: created}
			registry.RegistryServer.RegisterImage(env.Get("Id"), image)
			log.Printf("Register image id:%s host:%s", env.Get("Id"), c.Src)
		} else {
			log.Printf("Parse image created time error:%s", err)
		}
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
	config.Dockers[string(data)] = time.Now().Unix()
	log.Println("Docker:", string(data), "is online.")
	log.Println("Dockers:", config.Dockers)
	c.SendCommandString("docker_greetings_reply", config.ClusterAddress)
}

// 我收了个小弟
// todo:是否要审核
// 由于其知道我的名字我默认相信他
func dockerJoin(c *utils.Connection, data []byte) {
	// 向集群结构配置里面添加新成员
	config.Dockers[string(data)] = time.Now().Unix()
	// 返回组织中领导层所有人姓名以便小弟有事时着他们
	b, err := json.Marshal(config.Controllers)
	if err != nil {
		// 我收集的资料有误
		c.SendCommandString("docker_join", "")
	} else {
		// 告知其组织的领导层所有人员姓名
		c.SendCommandBytes("docker_join", b)
	}
}

// 结拜了个兄弟
func controllerJoin(c *utils.Connection, data []byte) {
	address := string(data)
	// 把他名字记下来
	config.Controllers[address] = time.Now().Unix()
	// 把我以前结拜的所有兄弟告诉他，让他们也认识一下
	b, err := json.Marshal(config.Controllers)
	if err != nil {
		log.Println(err)
		c.SendCommandString("controller_join", "")
	} else {
		c.SendCommandBytes("controller_join", b)
	}
	// 先断开连接，以免其收到广播我发给小弟的通知
	c.Conn.Close()
	// 告知我的所有小弟，我认了个兄弟，以后的进贡也要给他们一份

	message := fmt.Sprintf("%s %s", address, "controller_join_to_docker")
	ClusterSwitcher.Broadcast(utils.PacketString(message))
	log.Println("Controllers", config.Controllers)
}

// 小弟说我结拜的兄弟死了
func controllerOffline(c *utils.Connection, data []byte) {
	log.Println("controller:", string(data), "is offline.")
	// 从生死簿中将他的名字抹去
	delete(config.Controllers, string(data))
	log.Println("Controllers:", config.Controllers)
}

// 新的controller加入，docker需要连接到它
func controllerJoinToDocker(c *utils.Connection, data []byte) {
	address := string(data)
	log.Printf("Connect new controller %s", address)
	// docker连接到新的controller
	connCloseCh <- address
}

func dockerGreetingsReply(c *utils.Connection, data []byte) {
	config.Controllers[string(data)] = time.Now().Unix()
	reportImagesAndContainers(c)
}
