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

func RPC_Addr(me int, conf map[string]string) string {
	ip,ok := conf["n"+strconv.Itoa(me)]
		if !ok {
			fmt.Println("Failed to find IP :"+"n"+strconv.Itoa(me));
			panic(conf)
		}

	if conf["use_same_port"] == "true" {
		p,err := strconv.Atoi(conf["RPCport"])
		if err != nil {
			println("Failed to parse conf[port]")
			panic(err)
		}
		return ip+":"+strconv.Itoa(p)
	}

	p,err := strconv.Atoi(conf["RPC_port_n"+strconv.Itoa(me)])
		if err != nil {
			fmt.Println("Failed to parse :"+"RPC_port_n"+strconv.Itoa(me)+":"+conf["port_n"+strconv.Itoa(me)]);
			panic(err)
		}
	return ip+":"+strconv.Itoa(p)	
}

func main(){
	
	runtime.GOMAXPROCS(4)
	const nservers = 3
	var kva []*kvpaxos.KVPaxos = make([]*kvpaxos.KVPaxos, nservers)
	var kvh []string = make([]string, nservers)

	conf:=kvlib.ReadJson("conf/settings.conf")
	for i := 0; i < nservers; i++ {
		kvh[i] = RPC_Addr(i,conf)
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
