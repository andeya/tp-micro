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

func (w *Worker) Todo1(task string, reply *string) error {
	log.Println("Worker: do job", task)
	// time.Sleep(time.Second * 3)
	*reply = "OK"
	return nil
}

func (w *Worker) Todo2(task string, reply *string) error {
	log.Println("Worker: do job", task)
	// time.Sleep(time.Second * 3)
	*reply = "OK"
	return nil
}

// register itself as Plugin.
func (t *Worker) PostReadRequestHeader(req *rpc.Request) error {
	fmt.Printf("Worker4444.PostReadRequestHeader\n")
	return nil
}

func (t *Worker) PostReadRequestBody(_ interface{}) error {
	fmt.Printf("Worker4444.PostReadRequestBody\n")
	return nil
}

type testPlugin struct{}

func (t *testPlugin) PostReadRequestHeader(req *rpc.Request) error {
	s := rpc2.ParseServiceMethod(req.ServiceMethod)
	v, err := s.ParseQuery()
	if err != nil {
		return err
	}
	fmt.Printf("testPlugin1111.PostReadRequestHeader -> key[%d]: %#v\n", req.Seq, v.Get("key"))
	return nil
}

func (t *testPlugin) PostReadRequestBody(body interface{}) error {
	fmt.Printf("testPlugin1111.PostReadRequestBody -> %#v\n", body)
	return nil
}

type testPlugin2 struct{}

func (t *testPlugin2) PostReadRequestHeader(req *rpc.Request) error {
	fmt.Printf("testPlugin2222.PostReadRequestHeader\n")
	return nil
}

func (t *testPlugin2) PostReadRequestBody(body interface{}) error {
	fmt.Printf("testPlugin2222.PostReadRequestBody\n")
	return nil
}

type testPlugin3 struct{}

func (t *testPlugin3) PostReadRequestHeader(req *rpc.Request) error {
	fmt.Printf("testPlugin3333.PostReadRequestHeader\n")
	return nil
}

func (t *testPlugin3) PostReadRequestBody(body interface{}) error {
	fmt.Printf("testPlugin3333.PostReadRequestBody\n")
	return nil
}

// rpc2
func main() {
	// server
	server := rpc2.NewDefaultServer(true)
	server.IP().Allow("127.0.0.1")
	group, err := server.Group("test", new(testPlugin), new(testPlugin2))
	if err != nil {
		panic(err)
	}
	err = group.RegisterName("1.0.work", NewWorker(), new(testPlugin3))
	if err != nil {
		panic(err)
	}
	go server.ListenAndServe("tcp", "0.0.0.0:8080")
	time.Sleep(2e9)

	// client
	dialer := rpc2.NewDefaultDialer("tcp", "127.0.0.1:8080")
	client, _ := dialer.Dial()

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
	server.Stop()
	log.Println("cost time:", time.Now().Sub(t1))
	log.Println("failure rate:", float64(bad)/float64(good)*100, "%")
}

// standard package
func main2() {
	// server
	rpc.Register(NewWorker())
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", "127.0.0.1:8080") // any available address
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

				client, err := rpc.DialHTTP("tcp", "127.0.0.1:8080")
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
