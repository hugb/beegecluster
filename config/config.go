package config

var (
	Role string

	JoinAddress    string
	ServiceAddress string
	ClusterAddress string

	Dockers     = make(map[string]int64)
	Controllers = make(map[string]int64)
)
