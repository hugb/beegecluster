package cluster

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hugb/beegecontroller/config"
	"github.com/hugb/beegecontroller/utils"
)

func ClusterHandlers() {
	m := map[string]handlerFunc{
		"get_cluster_servers": GetClusterServers,
	}
	for cmd, fct := range m {
		if err := clusterSwitcher.Register(cmd, fct); err != nil {
			log.Printf("register cluster handler[%s] failure:%s\n", cmd, err)
		} else {
			log.Printf("register cluster hander[%s] success\n", cmd)
		}
	}
}

func GetClusterServers(c *utils.Connection, data []byte) {
	b, err := json.Marshal(config.CS)
	if err != nil {
		c.WriteFails(fmt.Sprintf("%s", err))
	} else {
		c.WriteSuccessBytes(b)
	}
}
