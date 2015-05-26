package main

import(
  "time"
  "strconv"
  "os"
  . "kvlib"
  )








func main(){
  conf := ReadJson("conf/settings.conf")
  rootURL :="http://"+conf["primary"]+":"+conf["port"]+"/"
  kvURL := rootURL+"kv/"
  //kvmanURL := rootURL+"kvman/"

  StartServer("-p");
  StartServer("-b");
  time.Sleep(500*time.Millisecond)
  N:=500
  if len(os.Args)>1 {
    N,_ = strconv.Atoi(os.Args[1])
  }

  TestPerformance(N, kvURL)

  StopServer("-p")
  StopServer("-b")
}
