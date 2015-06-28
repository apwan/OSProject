package main

import(
  //"net/http"
  "fmt"
  "os"
  "os/exec"
  "strconv"
  "bytes"

  // our lib
  . "kvlib"
  //. "paxos"
  //. "kvpaxos"
)

func usage(){
  fmt.Println("Usage: bin/stop_server <id>")
  os.Exit(1)
}

func main(){
  conf := ReadJson("conf/settings.conf")

  if len(os.Args)<=1 {
    usage()
  }else{
    if id,e := strconv.Atoi(os.Args[1]); e!=nil{
      fmt.Println(e)
      usage()
    }else{
      fmt.Printf("Stop Server %d\n", id)
      port := Find_Port(id-1, conf)
      cmd := exec.Command("lsof",[]string{"-t", "-i:"+strconv.Itoa(port)}...);
      o,_ := cmd.Output()
      if len(o)<=1 {
        fmt.Println("Fail to get pid")
        os.Exit(1)
      }
      pid := string(o[:bytes.IndexByte(o,'\n')])
      fmt.Printf("server pid: %s\n", pid)
      _,e = strconv.Atoi(pid)
      if e!=nil {
        fmt.Printf("Fail to parse pid: %s\n", e)
        os.Exit(1)
      }
      cmd = exec.Command("kill",[]string{pid}...)
      e = cmd.Run()
      if e!=nil {
        fmt.Printf("Stop Server Failed: %s\n",e)
        os.Exit(1)
      }

    }
  }

}
