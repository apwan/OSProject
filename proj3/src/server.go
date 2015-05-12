package main

import (
  "net"
  "net/url"
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
func find_port() (int,int){
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
		
	if conf["primary"]!=conf["backup"]{
		return p,p
	}
	
	if role==PRIMARY{
		return p,bp
	}
	return bp,bp
}
var(
 role = det_role() //PRIMARY, SECONDARY
 stage = COLD_START // COLD_START=0 WARM_START=1 BOOTSTRAP=2 SYNC=3
 conf = readConf("conf/settings.conf")
 listenPort, backupPort = find_port()
 db = cmap_string_string.New()
 )
 type BoolResponse struct {
    Success bool `json:"success"`
}
var (
	TrueResponseStr = "{\"success\":true}"
	FalseResponseStr = "{\"success\":false}"
)// in high-performance setting, TRS="1", FRS="0" !!!
 
 
var short_timeout = time.Duration(500 * time.Millisecond)
func dialTimeout(network, addr string) (net.Conn, error) {
    return net.DialTimeout(network, addr, short_timeout)
}
var fastTransport http.RoundTripper = &http.Transport{
        Proxy:                 http.ProxyFromEnvironment,
        ResponseHeaderTimeout: short_timeout,
		Dial: dialTimeout,
}
var fastClient = http.Client{
        Transport: fastTransport,
    }
var backup_furl = "http://"+conf["primary"]+":"+strconv.Itoa(backupPort)+"/kv/upsert"
func fastSync(key string, value string, del bool) bool{
	var url = backup_furl+
		"?key="+url.QueryEscape(key)+
		"&value="+url.QueryEscape(value)
	if(del){
		url += "&delete=true"
	}
	//key, value, delete=true
	resp, err := fastClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return string(body)=="1" ||  string(body)== TrueResponseStr
}
 
 
 //Main program started here
 //methods: insert,delete,update; get (via GET)


type StrResponse struct {
	Success bool `json:"success"`
    Value string `json:"value"`
}

func naive_kvUpsertHandler(w http.ResponseWriter, r *http.Request) {
	key:= r.FormValue("key")
	value:= r.FormValue("value")
	delete:= r.FormValue("delete")
	if(delete == "true"){
		db.Remove(key)
		fmt.Fprintf(w, "%s",TrueResponseStr)
		return
	}
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

func primary_kvGetHandler(w http.ResponseWriter, r *http.Request) {
	switch stage {
		case COLD_START, WARM_START:
			fmt.Fprintf(w, "%s",FalseResponseStr)
			return
		case BOOTSTRAP, SYNC:
			naive_kvGetHandler(w, r);
			return
	}
}

func primary_kvInsertHandler(w http.ResponseWriter, r *http.Request) {
	switch stage {
		case COLD_START, WARM_START, BOOTSTRAP:
			fmt.Fprintf(w, "%s",FalseResponseStr)
			return
	}
	//case SYNC:
	key:= r.FormValue("key")
	value:= r.FormValue("value")
	if !db.Has(key) && db.Set(key,value){
		ret:= fastSync(key,value,false)
		if ret{
			fmt.Fprintf(w, "%s",TrueResponseStr)
			return
		}
		//recover
		db.Remove(key)
	}
	fmt.Fprintf(w, "%s",FalseResponseStr)
}

func primary_kvUpdateHandler(w http.ResponseWriter, r *http.Request) {
}
func primary_kvDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	
	if role==BACKUP{// should be if(backup)
		http.HandleFunc("/kv/get", naive_kvGetHandler)
		http.HandleFunc("/kv/insert", naive_kvInsertHandler)
		http.HandleFunc("/kv/update", naive_kvUpdateHandler)
		http.HandleFunc("/kv/delete", naive_kvDeleteHandler)
		http.HandleFunc("/kv/upsert", naive_kvUpsertHandler)
	}else{
		http.HandleFunc("/kv/get", primary_kvGetHandler)
		http.HandleFunc("/kv/insert", primary_kvInsertHandler)
		http.HandleFunc("/kv/update", primary_kvUpdateHandler)
		http.HandleFunc("/kv/delete", primary_kvDeleteHandler)
	}
	log.Fatal(s.ListenAndServe())  
}
