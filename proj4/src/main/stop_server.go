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

var (
  role = Det_role()
)

func usage(){
  fmt.Println("Usage: bin/stop_server <n01|n02|...>")
  os.Exit(1)
}

func main(){
  conf := ReadJson("conf/settings.conf")

  if role<0 {
    usage()
  }else{

      fmt.Printf("Stop Server %d\n", role)
      port := Find_Port(role-1, conf)
      // lsof -t -i:[port]
      cmd := exec.Command("lsof",[]string{"-t", "-i:"+strconv.Itoa(port)}...);
      o,_ := cmd.Output()
      if len(o)<=1 {
        fmt.Println("Fail to get pid")
        os.Exit(1)
      }
      pid := string(o[:bytes.IndexByte(o,'\n')])
      fmt.Printf("Server PID: %s\n", pid)
      _,e := strconv.Atoi(pid)
      if e!=nil {
        fmt.Printf("Fail to parse PID: %s\n", e)
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
