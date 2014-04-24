package cluster

import (
	"log"
	"net"

	"github.com/hugb/beegecontroller/config"
	"github.com/hugb/beegecontroller/utils"
)

// controller监听docker的连接
func NewClusterServer() {
	var (
		err  error
		conn net.Conn
		ln   net.Listener
	)
	if ln, err = net.Listen("tcp", config.CS.ClusterAddress); err != nil {
		panic(err)
	}
	for {
		if conn, err = ln.Accept(); err == nil {
			go serve(conn)
		}
	}
}

// 从连接中读取数据，解析并调用相应handler响应
func serve(conn net.Conn) {
	var (
		exist      bool
		err        error
		cmd        string
		data       []byte
		payload    []byte
		length     int
		handler    HandlerFunc
		connection *utils.Connection
	)

	connection = &utils.Connection{Conn: conn}
	ClusterSwitcher.register <- connection

	defer func() {
		conn.Close()
		ClusterSwitcher.unregister <- connection
	}()

	for {
		if length, data, err = connection.Read(); err != nil {
			break
		}
		cmd, payload = utils.CmdDecode(length, data)
		log.Printf("Cmd:%s, payload:%s", cmd, string(payload))
		if handler, exist = ClusterSwitcher.handlers[cmd]; exist {
			handler(connection, payload)
		} else {
			connection.SendFailsResult(cmd, "Command does not exist")
		}
	}
}
