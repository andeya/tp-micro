package server

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/henrylee2cn/rpc2/gracenet"
	"github.com/henrylee2cn/rpc2/log"
	kcp "github.com/xtaci/kcp-go"
)

var (
	servers          []*Server
	serversLock      sync.RWMutex
	finalizers       []func() error
	SHUTDOWN_TIMEOUT = time.Minute
	shutdownTimeout  = SHUTDOWN_TIMEOUT
	graceSignalOnce  sync.Once
	exit             = make(chan bool)
)

func addServers(srvs ...*Server) {
	go graceSignalOnce.Do(graceSignal)
	serversLock.Lock()
	defer serversLock.Unlock()
	servers = append(servers, srvs...)
}

// SetShutdown sets the function which is called after the services shutdown,
// and the time-out period for the service shutdown.
// If parameter timeout is 0, automatically use default `SHUTDOWN_TIMEOUT`(60s).
// If parameter timeout less than 0, it is indefinite period.
// The finalizer function is executed before the shutdown deadline, but it is not guaranteed to be completed.
func SetShutdown(timeout time.Duration, fn ...func() error) {
	if timeout == 0 {
		timeout = SHUTDOWN_TIMEOUT
	} else if timeout < 0 {
		timeout = 1<<63 - 1
	}
	shutdownTimeout = timeout
	finalizers = fn
}

// Shutdown closes all the frame services gracefully.
// Parameter timeout is used to reset time-out period for the service shutdown.
func Shutdown(timeout ...time.Duration) {
	serversLock.Lock()
	defer serversLock.Unlock()
	log.Infof("shutting down servers...")
	if len(timeout) > 0 {
		SetShutdown(timeout[0], finalizers...)
	}
	graceful := shutdown()
	if graceful {
		log.Infof("servers are shutted down gracefully.")
	} else {
		log.Infof("servers are shutted down, but not gracefully.")
	}
}

func shutdown() bool {
	ctxTimeout, _ := context.WithTimeout(context.Background(), shutdownTimeout)
	count := new(sync.WaitGroup)
	var flag int32 = 1
	for _, server := range servers {
		count.Add(1)
		go func(svr *Server) {
			err := svr.close(ctxTimeout)
			if err != nil {
				log.Errorf("[shutdown-%s] %s", server.Address(), err.Error())
				atomic.StoreInt32(&flag, 0)
			}
			count.Done()
		}(server)
	}
	count.Wait()

	fchan := make(chan bool)
	var idx int32
	go func() {
		for i, finalizer := range finalizers {
			atomic.StoreInt32(&idx, int32(i))
			if finalizer == nil {
				continue
			}
			select {
			case <-ctxTimeout.Done():
				break
			default:
				if err := finalizer(); err != nil {
					atomic.StoreInt32(&flag, 0)
					log.Errorf("[shutdown-finalizer%d] %s", i, err.Error())
				}
			}
		}
		close(fchan)
	}()
	select {
	case <-ctxTimeout.Done():
		if err := ctxTimeout.Err(); err != nil {
			atomic.StoreInt32(&flag, 0)
			log.Errorf("[shutdown-finalizer%d] %s", atomic.LoadInt32(&idx), err.Error())
		}
	case <-fchan:
	}
	return flag == 1
}

var grace = new(gracenet.Net)

func makeListener(network, address string) (ln net.Listener, err error) {
	switch network {
	case "kcp":
		ln, err = kcp.ListenWithOptions(address, nil, 10, 3)
	default: //tcp
		ln, err = grace.Listen(network, address)
		// ln, err = net.Listen(network, address)
	}
	return ln, err
}
