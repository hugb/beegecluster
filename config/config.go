package config

var CS = &ClusterServer{}

type ClusterServer struct {
	docker     []string
	Controller []string
}
