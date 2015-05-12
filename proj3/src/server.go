package main

import (
  "net/http"
  "html"
  //"log"
  //"time"
  "fmt"
  "strconv"
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

const(
 PRIMARY=1
 BACKUP=2
)
const(
 COLD_START=0
 WARM_START=1
 BOOTSTRAP=2
 SYNC=3
)

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
func find_port() int{
	var p,err=strconv.Atoi(conf["port"])
		if err!=nil {
			fmt.Println("Failed to parse port:"+conf["port"]);
			panic(err)
		}
	var bp,err2=strconv.Atoi(conf["back_port"])
		if err!=nil{
			fmt.Println("Failed to parse back_port:"+conf["back_port"]);
			panic(err2)
		}
		
	if role==PRIMARY{
		return p
	}
	if conf["primary"]!=conf["backup"]{
		return p
	}
	return bp
}
var(
 role = det_role() //PRIMARY, SECONDARY
 stage = COLD_START // COLD_START=0 WARM_START=1 BOOTSTRAP=2 SYNC=3
 conf = readConf("conf/settings.conf")
 listenPort = find_port()
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
	fmt.Print("listenPort:")
	fmt.Println(listenPort)
	
	
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
