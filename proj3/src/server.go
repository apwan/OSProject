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
    //fmt.Println(dat)
	var udat map[string]interface{}
	if err := json.Unmarshal(dat, &udat); err != nil {
        panic(err)
    }
	ret:=make(map[string]string)
	
	for idx, val := range udat {
		var str=val.(string)
		ret[idx]=str
    }
    //fmt.Println("parsed config:")
	//fmt.Println(ret)
	return ret
}

const PRIMARY=1
const BACKUP=2
func det_role() int {
	arg_num := len(os.Args)
	for i := 0 ; i < arg_num ;i++{
		switch os.Args[i] {
			case "-p": 
				return PRIMARY
			case "-b": 
				return BACKUP
		}
	}
	return 0
}
  
var(
 role = det_role() //0 is uninitialized, 1 is primary, 2 is secondary
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
	fmt.Println("Initialized with conf:");
	fmt.Println(conf);
	
	fmt.Print("My role:");
	switch role{
		case PRIMARY:
			fmt.Println("Primary")
		case BACKUP:
			fmt.Println("Backup")
		default :
			fmt.Println("Unknown; please specify role as command line parameter.")
			panic(os.Args)
	}
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
	
}
