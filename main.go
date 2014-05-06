///////////////////////////////////////////////////////////////////
/*                          Controller                           */
///////////////////////////////////////////////////////////////////
// 提供API路由选择，以及集群管理
package main

import (
	"flag"

	"github.com/hugb/beegecluster/module"
)

func main() {
	var (
		joinAddress    = flag.String("j", "", "Join Address")
		serviceAddress = flag.String("p", "", "Service Address")
		clusterAddress = flag.String("c", "", "Cluster Address")
	)
	flag.Parse()

	// 启动控制器模块
	module.StartControllerrModule(*joinAddress, *serviceAddress, *clusterAddress)
}
