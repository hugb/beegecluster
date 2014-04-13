package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hugb/beegecontroller/config"
	"github.com/hugb/beegecontroller/utils"
)

func ControllerJoinCluster() {
	// 获取所有controller
	getController(config.CS.JoinPoint, "controller")

	log.Println("controllers", config.CS.ClusterServer.Controller)
}

func DockerJoinCluster() {
	// 获取所有controller
	getController(config.CS.JoinPoint, "docker")

	log.Println("controllers", config.CS.ClusterServer.Controller)

	// 连接所有controller
	for value, _ := range config.CS.ClusterServer.Controller {
		go connectController(value)
	}

	// todo:初始数据上报，包括镜像和容器

	// todo:docker服务器状态上报

	// todo:事件监听上报

	select {}

}

// docker连接到controller，保持着
// todo:reconnection，断开重试
func connectController(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	c := &utils.Connection{
		Conn: conn,
		Src:  address,
	}
	ClusterSwitcher.register <- c

	defer func() {
		conn.Close()
		ClusterSwitcher.unregister <- c
	}()

	c.SendCommandString("docker_greetings", config.CS.ServiceAddress)

	for {
		lenght, data, err := c.Read()
		if err != nil {
			break
		}

		cmd, code, payload := utils.CmdResultDecode(lenght, data)

		log.Printf("cmd:%s,code:%s,data:%s", cmd, code, string(payload))

		if handler, exist := ClusterSwitcher.handlers[cmd]; exist {
			handler(c, payload)
		}
	}
}

// 由入口地址得到所有的controller
func getController(address, from string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	c := &utils.Connection{
		Conn: conn,
	}

	defer func() {
		conn.Close()
	}()

	if from == "docker" {
		c.WriteString(fmt.Sprintf("%s_join_cluster", from), config.CS.ServiceAddress)
	}
	if from == "controller" {
		c.WriteString(fmt.Sprintf("%s_join_cluster", from), config.CS.ClusterAddress)
	}

	lenght, data, err := c.Read()
	if err != nil {
		return
	}

	cmd, code, payload := utils.CmdResultDecode(lenght, data)
	log.Printf("cmd:%s, code:%s, payload:%s", cmd, code, string(payload))
	if code == utils.FAILURE {
		return
	}

	var controllers map[string]int64
	if err = json.Unmarshal(payload, &controllers); err != nil {
		log.Println("decode json error:", err)
	}

	for address, _ := range controllers {
		if _, exist := config.CS.ClusterServer.Controller[address]; !exist {
			config.CS.ClusterServer.Controller[address] = time.Now().Unix()
			getController(address, from)
		}
	}
}
