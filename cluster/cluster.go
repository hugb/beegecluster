package cluster

import (
	"net"

	"github.com/hugb/beegecontroller/utils"
)

func NewClusterServer(address string) {
	ln, err := net.Listen("tcp", address)
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

	defer func() {
		clusterSwitcher.unregister <- c
		conn.Close()
	}()

	clusterSwitcher.register <- c

	for {
		cmd, data, err := c.Read()
		if err != nil {
			break
		}

		if handler, exist := clusterSwitcher.handlers[cmd]; exist {
			handler(c, data)
		} else {
			c.WriteFails("Command does not exist")
		}
	}
}
