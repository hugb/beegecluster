package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	gosignal "os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dotcloud/docker/engine"
	dockerUtils "github.com/dotcloud/docker/utils"

	"github.com/hugb/beegecluster/config"
	"github.com/hugb/beegecluster/utils"
)

var (
	Eng *engine.Engine

	waitGroup sync.WaitGroup

	connCloseCh chan string
)

// 我要加入组织
func ControllerJoinCluster() {
	// 得到各个分社的领导人姓名
	getController(config.JoinAddress)
	// 所有领导人
	log.Println("Controllers:", config.Controllers)
}

func DockerJoinCluster(eng *engine.Engine) {
	Eng = eng
	// 捕获系统信号
	c := make(chan os.Signal, 1)
	gosignal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range c {
			// os.Interrupt=ctrl+c
			log.Printf("Received signal '%v'", sig)
		}
	}()

	// 获取所有controller的集群内部通信地址
	getController(config.JoinAddress)

	log.Println("Controllers:", config.Controllers)

	connCloseCh = make(chan string, 10)

	// 连接所有controller
	for address, _ := range config.Controllers {
		waitGroup.Add(1)
		go connectController(address)
	}
	// 等待docker与controller所有连接完成
	waitGroup.Wait()

	// 监听到连接断开进行重连
	go reConnectController()

	// 每月还贷
	go reportStatus()

	// 事件上报
	if err := reportEvents(); err != nil {
		log.Println("Report event error:", err)
	}

	log.Println("Event report finish")
}

func reportImagesAndContainers(c *utils.Connection) error {
	// 读取镜像列表
	log.Println("Report images start.")
	imageJob := Eng.Job("images")
	imageJob.Setenv("filter", "")
	imageJob.Setenv("all", "0")
	imageSrc, err := imageJob.Stdout.AddPipe()
	if err != nil {
		log.Fatalf("Create images receive pipe error:%s", err)
	}
	// 从管道读取事件数据并广播给所有controller
	go func() {
		imagesBytes, err := ioutil.ReadAll(imageSrc)
		if err != nil {
			log.Println("Read data error from pipe:", err)
		}
		c.SendCommandBytes("docker_images", imagesBytes)
	}()
	if err := imageJob.Run(); err != nil {
		return err
	}
	// 读取容器列表
	log.Println("Report containers start")
	containerJob := Eng.Job("containers")
	containerJob.Setenv("all", "1")
	containerSrc, err := containerJob.Stdout.AddPipe()
	if err != nil {
		log.Fatalf("Create containers receive pipe error:%s", err)
	}
	// 从管道读取事件数据并广播给所有controller
	go func() {
		containersBytes, err := ioutil.ReadAll(containerSrc)
		if err != nil {
			log.Println("Read data error from pipe:", err)
		}
		c.SendCommandBytes("docker_containers", containersBytes)
	}()

	return containerJob.Run()
}

// 上报docker事件
func reportEvents() error {
	job := Eng.Job("events", "DockerAgent")
	// 从当前到3214080000（100年）后,^-^100后我都不在了，还需要考虑超时吗
	job.Setenv("since", fmt.Sprint(time.Now().Unix()))
	job.Setenv("until", fmt.Sprint(time.Now().Unix()+3214080000))
	reader, err := job.Stdout.AddPipe()
	if err != nil {
		log.Fatalf("Create event receive pipe error:%s", err)
	}
	// 从管道读取事件数据并广播给所有controller
	go func() {
		dec := json.NewDecoder(reader)
		for {
			m := &dockerUtils.JSONMessage{}
			if err := dec.Decode(m); err != nil {
				log.Printf("Error streaming events: %s\n", err)
				break
			}
			if b, err := json.Marshal(m); err == nil {
				// 广播
				log.Println("Event:", string(b))
				content := utils.PacketByes(append(b, " docker_event"...))
				ClusterSwitcher.Broadcast(content)
			}
		}
	}()

	return job.Run()
}

// 上报docker主机状态
func reportStatus() {
	log.Println("Report status start")
	tick := time.Tick(time.Duration(5) * time.Second)
	for {
		select {
		case <-tick:
			systemInfo, err := utils.GetSystemInfo()
			if err != nil {
				log.Println("Get system info error:", err)
			}
			// 包含cpu使用率，内存，交换区和负载信息
			systemInfoBytes, err := json.Marshal(systemInfo)
			if err != nil {
				log.Println("Encode system info error:", err)
			}
			log.Println("Report status...")
			data := utils.PacketByes(append(systemInfoBytes, " docker_status"...))
			ClusterSwitcher.Broadcast(data)
		}
	}
	log.Println("Report status finish")
}

// docker连接到controller，保持着
func connectController(address string) {
	defer func() { connCloseCh <- address }()
	var (
		length     int
		exist      bool
		err        error
		cmd        string
		data       []byte
		payload    []byte
		conn       net.Conn
		handler    HandlerFunc
		connection *utils.Connection
	)

	log.Printf("Connect controller %s", address)

	conn, err = net.Dial("tcp", address)
	if err != nil {
		log.Printf("Connect controller %s error:%s", address, err)
		return
	}
	connection = &utils.Connection{
		Conn: conn,
		Src:  address,
	}
	ClusterSwitcher.register <- connection

	defer func() {
		conn.Close()
		ClusterSwitcher.unregister <- connection
	}()

	waitGroup.Done()

	connection.SendCommandString("docker_greetings", config.ClusterAddress)

	for {
		if length, data, err = connection.Read(); err != nil {
			break
		}
		cmd, payload = utils.CmdDecode(length, data)

		log.Printf("Cmd:%s", cmd)

		if handler, exist = ClusterSwitcher.handlers[cmd]; exist {
			handler(connection, payload)
		} else {
			log.Printf("Command[%s] is not exist", cmd)
		}
	}
	log.Printf("Controller %s is disconnect", address)
}

// 由入口地址得到所有的controller
func getController(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	connection := &utils.Connection{Conn: conn}

	defer func() { conn.Close() }()

	log.Println("Get all controllers request")
	connection.SendCommandString(fmt.Sprintf("%s_join", config.Role), config.ClusterAddress)

	lenght, data, err := connection.Read()
	if err != nil {
		panic(err)
	}

	cmd, payload := utils.CmdDecode(lenght, data)

	log.Printf("Response cmd:%s, payload:%s", cmd, string(payload))

	var controllers map[string]int64
	if err = json.Unmarshal(payload, &controllers); err != nil {
		panic(err)
	}

	for address, _ := range controllers {
		if _, exist := config.Controllers[address]; !exist {
			config.Controllers[address] = time.Now().Unix()
			getController(address)
		}
	}
}

// 重新连接到Controller
func reConnectController() {
	var (
		ok      bool
		address string
	)
	for address = range connCloseCh {
		// 广播控制器已经离线
		//log.Printf("Broadcast controller[%s] is offline", address)
		//message := fmt.Sprintf("%s %s", address, "controller_offline")
		//ClusterSwitcher.Broadcast(utils.PacketString(message))

		if _, ok = config.Controllers[address]; ok {
			continue
		}

		for i := 3; i > 0; i-- {
			log.Printf("Wait %d seconed reconnect %s", i, address)
			time.Sleep(1 * time.Second)
		}

		waitGroup.Add(1)
		go connectController(address)
		waitGroup.Wait()
	}
}
