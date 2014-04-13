package cluster

import (
	"fmt"
	"log"
	"time"

	"github.com/hugb/beegecontroller/utils"
)

const (
	maxMessageSize = 256
)

type HandlerFunc func(c *utils.Connection, data []byte)

type Switcher struct {
	broadcast   chan []byte
	handlers    map[string]HandlerFunc
	register    chan *utils.Connection
	unregister  chan *utils.Connection
	connections map[*utils.Connection]int64
}

// 数据交换器
var ClusterSwitcher = &Switcher{
	handlers:    make(map[string]HandlerFunc),
	register:    make(chan *utils.Connection, 1),
	unregister:  make(chan *utils.Connection, 1),
	connections: make(map[*utils.Connection]int64),
	broadcast:   make(chan []byte, maxMessageSize),
}

func init() {
	// 启动交换器
	go ClusterSwitcher.run()
}

// 连接管理以及数据广播
func (this *Switcher) run() {
	for {
		select {
		case c := <-this.register:
			this.connections[c] = time.Now().Unix()
		case c := <-this.unregister:
			if c.Src != "" {
				if handler, exist := this.handlers["disconnect"]; exist {
					handler(c, []byte(c.Src))
				}
			}
			delete(this.connections, c)
		case m := <-this.broadcast:
			for c := range this.connections {
				c.Conn.Write(m)
			}
		}
	}
}

// 向制定的docker发送数据
func (this *Switcher) Unicast(address string, data []byte) {
	for conn, _ := range this.connections {
		// todo:只有docker的连接src才不为空
		if conn.Src == address {
			conn.SendSuccessResultBytes("unicast", data)
		}
	}
}

// 向制定的docker发送数据
func (this *Switcher) Broadcast(data []byte) {
	this.broadcast <- data
}

func (this *Switcher) Register(command string, handler HandlerFunc) error {
	if _, exists := this.handlers[command]; exists {
		return fmt.Errorf("Can't overwrite handler for command %s", command)
	}
	this.handlers[command] = handler
	return nil
}
