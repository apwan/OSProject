package main

import(
  "net/http"
  "fmt"
  )

func main(){
  fmt.Println("test kv")
  resp, err := http.Get("localhost:8088/kv")
  if err != nil{
    fmt.Println(err)
  }else{
    fmt.Println(resp);
  }

  fmt.Println("test kvman")
  resp, err = http.Get("localhost:8088/kvman")
  if err != nil{
    fmt.Println(err)
  }else{
    fmt.Println(resp);
  }
}
