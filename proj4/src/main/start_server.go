package main

import(
	//"net/http"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"runtime"
	"strconv"

	// our lib
	//. "kvlib"
	//. "paxos"
	"kvpaxos"
)

func port(tag string, host int) string {
  s := "/var/tmp/824-"
  s += strconv.Itoa(os.Getuid()) + "/"
  os.Mkdir(s, 0777)
  s += "kv-"
  s += strconv.Itoa(os.Getpid()) + "-"
  s += tag + "-"
  s += strconv.Itoa(host)
  return s
}

func main(){
	//conf := ReadJson("conf/settings.conf")
	//fmt.Println(conf["port"])

	runtime.GOMAXPROCS(4)
	const nservers = 3
	var kva []*kvpaxos.KVPaxos = make([]*kvpaxos.KVPaxos, nservers)
	var kvh []string = make([]string, nservers)

	for i := 0; i < nservers; i++ {
		kvh[i] = port("basic", i)
	}
	for i := 0; i < nservers; i++ {
		kva[i] = kvpaxos.StartServer(kvh, i)
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT)
	fmt.Printf("Serving HTTP\n")
	select {
		case signal := <-stop:
			fmt.Printf("Got signal:%v\n", signal)

		for i := 0; i < nservers; i++ {
			kva[i].Kill()
		}
	}

}
