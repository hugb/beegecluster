package cluster

import (
	"fmt"
	"time"

	"github.com/hugb/beegecontroller/utils"
)

const (
	maxMessageSize = 256
)

type handlerFunc func(c *utils.Connection, data []byte)

type switcher struct {
	broadcast   chan []byte
	handlers    map[string]handlerFunc
	register    chan *utils.Connection
	unregister  chan *utils.Connection
	connections map[*utils.Connection]int64
}

var clusterSwitcher = &switcher{
	handlers:    make(map[string]handlerFunc),
	register:    make(chan *utils.Connection, 1),
	unregister:  make(chan *utils.Connection, 1),
	connections: make(map[*utils.Connection]int64),
	broadcast:   make(chan []byte, maxMessageSize),
}

func init() {
	go clusterSwitcher.run()
}

func (this *switcher) run() {
	for {
		select {
		case c := <-this.register:
			this.connections[c] = time.Now().Unix()
		case c := <-this.unregister:
			delete(this.connections, c)
		case m := <-this.broadcast:
			for c := range this.connections {
				c.WriteSuccessBytes(m)
			}
		}
	}
}

func (this *switcher) Register(command string, handler handlerFunc) error {
	if _, exists := this.handlers[command]; exists {
		return fmt.Errorf("Can't overwrite handler for command %s", command)
	}
	this.handlers[command] = handler
	return nil
}
