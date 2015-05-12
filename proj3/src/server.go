package main

import (
  "net/http"
  "html"
  //"log"
  //"time"
  "fmt"
  "os"
  "io/ioutil"
  "encoding/json"
  )
 
  
func readConf(s string) map[string]string{
    dat, err := ioutil.ReadFile(s)
	if err != nil {
        panic(err)
    }
    fmt.Println(dat)
	var udat map[string]interface{}
	if err := json.Unmarshal(dat, &udat); err != nil {
        panic(err)
    }
    fmt.Println(udat)
	
	var ret map[string]string{}
	
	for idx, val := range udat {
        if t.(type) == string {
            ret[idx]=val
        }
    }
	return ret
}
  
var(
 role = 0 //0 is uninitialized, 1 is primary, 2 is secondary
 stage = 0 //0: cold start, 1: warm start, 2: bootstrap, 3: synced  
 conf = readConf("conf/settings.conf")
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
  /*s := &http.Server{
    Addr: ":8088",
    Handler: nil,
    ReadTimeout: 10 * time.Second,
    WriteTimeout: 10 * time.Second,
    MaxHeaderBytes: 1<<20,
  }
  http.HandleFunc("/kv", kvHandler)
  http.HandleFunc("/kvman", kvmanHandler)
  log.Fatal(s.ListenAndServe())*/
	arg_num := len(os.Args)
	fmt.Printf("the num of input is %d\n",arg_num)

	fmt.Printf("they are :\n")
	for i := 0 ; i < arg_num ;i++{
		fmt.Println(os.Args[i])
	}
}
