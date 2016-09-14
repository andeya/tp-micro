package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"github.com/henrylee2cn/rpc2"
)

type Worker struct {
	Name string
}

func NewWorker() *Worker {
	return &Worker{"test"}
}

func (w *Worker) DoJob(task string, reply *string) error {
	// log.Println("Worker: do job", task)
	// time.Sleep(time.Second * 3)
	*reply = "OK"
	return nil
}

// rpc2
func main() {
	// server
	rpc2.Register(NewWorker())
	rpc2.AddAllowedIPPrefix("127.0.0.1")
	go rpc2.ListenRPC("0.0.0.0:80")
	time.Sleep(2e9)

	// client
	N := 1000
	bad := 0
	good := 0
	mapChan := make(chan int, N)
	t1 := time.Now()
	for ii := 0; ii < 10; ii++ {
		for i := 0; i < N; i++ {
			go func(i int) {
				var reply = new(string)
				e := rpc2.Call("127.0.0.1:80", "Worker.DoJob", strconv.Itoa(ii*N+i), reply)
				log.Println(ii*N+i, *reply, e)
				if e != nil {
					mapChan <- 0
				} else {
					mapChan <- 1
				}
			}(i)
		}
		for i := 0; i < N; i++ {
			if r := <-mapChan; r == 0 {
				bad++
			} else {
				good++
			}
		}
	}
	log.Println("cost time:", time.Now().Sub(t1))
	log.Println("failure rate:", float64(bad)/float64(good)*100, "%")
}

// standard package
func main2() {
	// server
	rpc.Register(NewWorker())
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", "127.0.0.1:80") // any available address
	if e != nil {
		log.Fatalf("net.Listen tcp :0: %v", e)
	}

	go http.Serve(l, nil)
	time.Sleep(2e9)

	// client
	N := 1000
	bad := 0
	good := 0
	mapChan := make(chan int, N)
	t1 := time.Now()
	for ii := 0; ii < 10; ii++ {
		for i := 0; i < N; i++ {
			go func(i int) {
				var reply = new(string)
				var err error
				defer func() {
					if err != nil {
						mapChan <- 0
					} else {
						mapChan <- 1
					}
					log.Println(ii*N+i, *reply, err)
				}()

				client, err := rpc.DialHTTP("tcp", "127.0.0.1:80")
				if err != nil {
					log.Println("dialing:", err)
					return
				}

				err = client.Call("Worker.DoJob", strconv.Itoa(ii*N+i), reply)
				if err != nil {
					log.Println("arith error:", err)
					return
				}

			}(i)
		}
		for i := 0; i < N; i++ {
			if r := <-mapChan; r == 0 {
				bad++
			} else {
				good++
			}
		}
	}
	log.Println("cost time:", time.Now().Sub(t1))
	log.Println("failure rate:", float64(bad)/float64(good)*100, "%")

}
