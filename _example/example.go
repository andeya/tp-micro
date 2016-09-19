package main

import (
	"fmt"
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
	log.Println("Worker: do job", task)
	// time.Sleep(time.Second * 3)
	*reply = "OK"
	return nil
}

type testPlugin struct{}

func (t *testPlugin) PostReadRequestHeader(req *rpc.Request) error {
	s := rpc2.ParseServiceMethod(req.ServiceMethod)
	v, err := s.ParseQuery()
	if err != nil {
		return err
	}
	fmt.Printf("PostReadRequestHeader -> key[%d]: %#v\n", req.Seq, v.Get("key"))
	return nil
}

func (t *testPlugin) PostReadRequestBody(body interface{}) error {
	fmt.Printf("PostReadRequestBody -> %#v\n", body)
	return nil
}

// rpc2
func main() {
	// server
	rpc2.AddAllowedIPPrefix("127.0.0.1")
	server := rpc2.NewDefaultServer()
	g := server.Group("test", new(testPlugin))
	err := g.RegisterName("1.0.work", NewWorker())
	if err != nil {
		panic(err)
	}
	go server.ListenTCP("0.0.0.0:8080")
	time.Sleep(2e9)

	// client
	client := rpc2.NewClient("127.0.0.1:8080", nil)

	N := 1000
	bad := 0
	good := 0
	mapChan := make(chan int, N)
	t1 := time.Now()
	for ii := 0; ii < 10; ii++ {
		for i := 0; i < N; i++ {
			go func(i int) {
				var reply = new(string)
				// e := client.Call("Worker.DoJob", strconv.Itoa(ii*N+i), reply)
				e := client.Call("test/1.0.work.DoJob?key=henrylee2cn", strconv.Itoa(ii*N+i), reply)
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
	client.Close()
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
