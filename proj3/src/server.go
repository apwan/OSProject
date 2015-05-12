package main

import (
  "net/http"
  "html"
  "log"
  "time"
  "fmt"
  "os"
  "encoding/json"
  )

func kvHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q, we should implements insert/delete/get/update",
      html.EscapeString(r.URL.Path))
}
func kvmanHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q, we should implements countkey/dump/shutdown",
      html.EscapeString(r.URL.Path))
}
func main(){
  s := &http.Server{
    Addr: ":8088",
    Handler: nil,
    ReadTimeout: 10 * time.Second,
    WriteTimeout: 10 * time.Second,
    MaxHeaderBytes: 1<<20,
  }
  http.HandleFunc("/kv", kvHandler)
  http.HandleFunc("/kvman", kvmanHandler)
  log.Fatal(s.ListenAndServe())

}
