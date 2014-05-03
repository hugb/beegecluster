//////////////////////////////////////////////////////////
/* 监听一个端口，等待docker链接，完成controller与docker的通信 */
//////////////////////////////////////////////////////////

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
	panic("Cluster communication server stops")
}

// 从连接中读取数据，解析并调用相应handler响应
func serve(conn net.Conn) {
	var (
		n          int
		ok         bool
		err        error
		cmd        string
		data       []byte
		payload    []byte
		handler    HandlerFunc
		connection *utils.Connection
	)

	connection = &utils.Connection{Conn: conn}
	ClusterSwitcher.register <- connection

	defer func() {
		ClusterSwitcher.unregister <- connection
		conn.Close()
	}()

	for {
		if n, data, err = connection.Read(); err != nil {
			break
		}
		cmd, payload = utils.CmdDecode(n, data)

		log.Printf("Controller receive Cmd:%s, payload:%s", cmd, string(payload))

		if handler, ok = ClusterSwitcher.handlers[cmd]; ok {
			handler(connection, payload)
		} else {
			connection.SendFailsResult(cmd, "Command does not exist")
		}
	}
}
