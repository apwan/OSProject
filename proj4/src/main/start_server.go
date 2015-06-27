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
	"kvlib"
	//. "paxos"
	"kvpaxos"
)

var conf = kvlib.ReadJson("conf/settings.conf")

func RPCport(N int, me int) string {
	return "127.0.0.1:"+strconv.Itoa(40000+me)
}

func main(){
	
	runtime.GOMAXPROCS(4)
	const nservers = 3
	var kva []*kvpaxos.KVPaxos = make([]*kvpaxos.KVPaxos, nservers)
	var kvh []string = make([]string, nservers)

	for i := 0; i < nservers; i++ {
		kvh[i] = RPCport(nservers,i)
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
