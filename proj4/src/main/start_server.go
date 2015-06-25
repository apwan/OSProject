package main

import(
  //"net/http"
  "fmt"
  //"strconv"

  // our lib
  . "kvlib"
  //. "paxos"
  //. "kvpaxos"
)

func main(){
  conf := ReadJson("conf/settings.conf")
  fmt.Println(conf["port"])

}
