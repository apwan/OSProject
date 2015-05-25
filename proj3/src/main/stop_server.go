package main

// stop_server -p|-b

import(
  "net/http"
  "fmt"
  "strconv"
  . "kvlib"
  )

func main(){
  conf := ReadJson("conf/settings.conf")
  role := Det_role()
  port,_,_ := Find_port(role, conf)
  ipaddr := conf["primary"]
  if role==BACKUP {
    ipaddr = conf["backup"]
  }
  resp, err := http.Get("http://" +ipaddr+ ":" + strconv.Itoa(port) + "/kvman/shutdown")
  if err != nil{
    fmt.Println(err)
  }else{
    fmt.Println(DecodeStr(resp))
  }
}
