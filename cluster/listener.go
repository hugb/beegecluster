package cluster

import (
	"log"
	"net"

	"github.com/hugb/beegecontroller/config"
	"github.com/hugb/beegecontroller/utils"
)

// controller监听docker的连接
func NewClusterServer() {
	ln, err := net.Listen("tcp", config.CS.ClusterAddress)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go serve(conn)
	}
}

func serve(conn net.Conn) {
	c := &utils.Connection{
		Conn: conn,
	}
	ClusterSwitcher.register <- c

	defer func() {
		conn.Close()
		ClusterSwitcher.unregister <- c
	}()

	for {
		lenght, data, err := c.Read()
		if err != nil {
			break
		}

		cmd, payload := utils.CmdDecode(lenght, data)
		log.Printf("cmd:%s, payload:%s.", cmd, string(payload))

		if handler, exist := ClusterSwitcher.handlers[cmd]; exist {
			handler(c, payload)
		} else {
			c.SendFailsResult(cmd, "Command does not exist.")
		}
	}
}
