package rpc2

var servers []*Server

func addServers(srvs ...*Server) {
	servers = append(servers, srvs...)
}

// Close all servers
func Close() {
	for _, srv := range servers {
		srv.Close()
	}
}
