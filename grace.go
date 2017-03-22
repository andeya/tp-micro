package rpc2

// var servers = map[*Server]bool{}

// func addServers(srvs ...*Server) {
// 	for _, srv := range srvs {
// 		servers[srv] = true
// 	}
// }

// // Close all servers
// func Close() {
// 	for srv := range servers {
// 		srv.Close()
// 	}
// }

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
