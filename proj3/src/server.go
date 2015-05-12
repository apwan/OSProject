package main

import (
  "net/http"
  "html"
  "log"
  "time"
  "fmt"
  "strconv"
  "os"
  "io/ioutil"
  "encoding/json"
  "./cmap_string_string"
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
 db = cmap_string_string.New()
 )
 
 //Main program started here
 //methods: insert,delete,update; get (via GET)
type BoolResponse struct {
    Success bool `json:"success"`
}
var (
	TrueResponseStr = "{\"success\":true}"
	FalseResponseStr = "{\"success\":false}"
)// in high-performance setting, TRS="1", FRS="0" !!!

type StrResponse struct {
	Success bool `json:"success"`
    Value string `json:"value"`
}

func naive_kvUpsertHandler(w http.ResponseWriter, r *http.Request) {
	key:= r.FormValue("key")
	value:= r.FormValue("value")
	if db.Set(key,value){
		fmt.Fprintf(w, "%s",TrueResponseStr)
		return
	}
	fmt.Fprintf(w, "%s",FalseResponseStr)
} 
func naive_kvInsertHandler(w http.ResponseWriter, r *http.Request) {
	key:= r.FormValue("key")
	value:= r.FormValue("value")
	if !db.Has(key) && db.Set(key,value){
		fmt.Fprintf(w, "%s",TrueResponseStr)
		return
	}
	fmt.Fprintf(w, "%s",FalseResponseStr)
} 
func naive_kvUpdateHandler(w http.ResponseWriter, r *http.Request) {
	key:= r.FormValue("key")
	value:= r.FormValue("value")
	if db.Has(key) && db.Set(key,value){
		fmt.Fprintf(w, "%s",TrueResponseStr)
		return
	}
	fmt.Fprintf(w, "%s",FalseResponseStr)
} 
func naive_kvDeleteHandler(w http.ResponseWriter, r *http.Request) {
	key:= r.FormValue("key")
	if db.Has(key){
		db.Remove(key)
		fmt.Fprintf(w, "%s",TrueResponseStr)
		return
	}
	fmt.Fprintf(w, "%s",FalseResponseStr)
} 
func naive_kvGetHandler(w http.ResponseWriter, r *http.Request) {
	key:= r.FormValue("key")
	val,ok:= db.Get(key)
	ret:=&StrResponse{
		Success:ok,
		Value:val}
	str,_:=json.Marshal(ret);
	fmt.Fprintf(w, "%s",str)
} 
 
func kvmanCountkeyHandler(w http.ResponseWriter, r *http.Request) {
	tmp := make(map[string]int)
	tmp["result"]=db.Count()
	var str,err=json.Marshal(tmp)
	if err==nil{
		fmt.Fprintf(w, "%s",str)
		return
	}
	fmt.Fprintf(w, "DB marshalling error %s",err)
}
func kvmanDumpHandler(w http.ResponseWriter, r *http.Request) {
	var str,err=db.MarshalJSON();
	if err==nil{
		fmt.Fprintf(w, "%s",str)
		return
	}
	fmt.Fprintf(w, "DB marshalling error %s",err)
}
func kvmanShutdownHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q, DB suicide",
      html.EscapeString(r.URL.Path))
	go func(){
		time.Sleep(1) //sleep epsilon
		os.Exit(0)
	}()
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/"{
		time.Sleep(3 * time.Second)
		fmt.Fprintf(w, "Unrecognized Request")
		return
	}
	fmt.Fprintf(w, "Hello, %q, this is a server.",
      html.EscapeString(r.URL.Path))
	fmt.Fprintf(w, "Role:%d, stage:%d, dump:",
      role, stage)
	kvmanDumpHandler(w,r);
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
	
	
	db.Set("_","__");
  
	s := &http.Server{
		Addr: ":"+strconv.Itoa(listenPort),
		Handler: nil,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1<<20,
	}
	//http.HandleFunc("/kv", kvHandler)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/kvman/countkey", kvmanCountkeyHandler)
	http.HandleFunc("/kvman/dump", kvmanDumpHandler)
	http.HandleFunc("/kvman/shutdown", kvmanShutdownHandler)
	
	if true{// should be if(backup)
		http.HandleFunc("/kv/get", naive_kvGetHandler)
		http.HandleFunc("/kv/insert", naive_kvInsertHandler)
		http.HandleFunc("/kv/update", naive_kvUpdateHandler)
		http.HandleFunc("/kv/delete", naive_kvDeleteHandler)
		http.HandleFunc("/kv/upsert", naive_kvUpsertHandler)
	}
	log.Fatal(s.ListenAndServe())  
}
