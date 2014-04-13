package config

type ClusterServer struct {
	Docker     map[string]int64
	Controller map[string]int64
}

var defaultClusterServer = ClusterServer{
	Docker:     make(map[string]int64),
	Controller: make(map[string]int64),
}

var CS = &Config{
	ClusterServer: defaultClusterServer,
}

type Config struct {
	JoinPoint      string
	ServiceAddress string
	ClusterAddress string

	ClusterServer ClusterServer
}
