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
	id := fmt.Sprintf("n%02d", me+1)
	ip,ok := conf[id]
		if !ok {
			fmt.Println("Failed to find IP :"+id);
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

	p,err := strconv.Atoi(conf["RPC_port_"+id])
		if err != nil {
			fmt.Printf("Failed to parse : RPC_port_%s : %s\n",id, conf["port_"+id]);
			panic(err)
		}
	return ip+":"+strconv.Itoa(p)
}

func usage(){
	fmt.Println("Usage: bin/start_server <id>")
	os.Exit(1)
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

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT)

	if len(os.Args)>1 {
		fmt.Printf("single start mode: %s\n",os.Args[0])
		var kva_me *kvpaxos.KVPaxos
		if id,e := strconv.Atoi(os.Args[1]); e!=nil{
			usage()
		}else{
			kva_me = kvpaxos.StartServer(kvh, id-1)
			fmt.Printf("Serving HTTP, Server ID: %d\n", id)
		}
		select {
			case signal := <-stop:
				fmt.Printf("Got signal:%v\n", signal)
				kva_me.Kill()
		}

	}else{
		for i := 0; i < nservers; i++ {
			kva[i] = kvpaxos.StartServer(kvh, i)
		}
		fmt.Printf("Serving HTTP\n")
		select {
			case signal := <-stop:
				fmt.Printf("Got signal:%v\n", signal)

			for i := 0; i < nservers; i++ {
				kva[i].Kill()
			}
		}
	}
}
