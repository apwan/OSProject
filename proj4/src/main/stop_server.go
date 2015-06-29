package main

import(
  //"net/http"
  "fmt"
  "os/exec"
  "strconv"
  "bytes"
  "net/http"

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
}

func main(){
  if role<0 {
    usage()
    return
  }
  conf := ReadJson("conf/settings.conf")
  ips := []string{"127.0.0.1",conf["n01"],conf["n02"],conf["n03"]}
  ports := []string{conf["port"],conf["port_n01"],conf["port_n02"],conf["port_n03"]}
  if conf["use_different_port"]!="true"{
    ports[1],ports[2],ports[3] = ports[0],ports[0],ports[0]
  }

      fmt.Printf("Stop Server %d\n", role)
      addr := fmt.Sprintf("http://%s:%s/kvman/shutdown", ips[role], ports[role])

        resp,err := http.Get(addr)
        if err==nil{
          res:=DecodeStr(resp)
          fmt.Println(res)
          return
        }
          resp,err=http.Get(addr)
          if err==nil{
            res:=DecodeStr(resp)
            fmt.Println(res)
            return
          }
           // forced // lsof -t -i:[port]
            cmd := exec.Command("lsof",[]string{"-t", "-i:"+ports[role]}...);
            o,_ := cmd.Output()
            if len(o)<=1 {
              fmt.Println("Fail to get pid")
              return
            }
            pid := string(o[:bytes.IndexByte(o,'\n')])
            fmt.Printf("Server PID: %s\n", pid)
            _,e := strconv.Atoi(pid)
            if e!=nil {
              fmt.Printf("Fail to parse PID: %s\n", e)
              return
            }
            cmd = exec.Command("kill",[]string{pid}...)
            e = cmd.Run()
            if e!=nil {
              fmt.Printf("Stop Server Failed: %s\n",e)
              return
            }


}
