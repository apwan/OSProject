package main

import(
  "net"
  "fmt"
  "time"
  "strconv"
  "sort"
  "os"
  "strings"
  "regexp"
  . "kvlib"
  )


var conf = ReadJson("conf/settings.conf")

var(
 //rootURL="http://"+conf["primary"]+":"+conf["port"]+"/"
 //kvURL=rootURL+"kv/"
 //kvmanURL=rootURL+"kvman/"
 host=conf["primary"]+":"+conf["port"]
)
/*
func naive_HTTP(url string, data_enc string, post bool) (string, error) {
	if post{
		resp, err := http.Post(url,
			"application/x-www-form-urlencoded",
			strings.NewReader(data_enc))
		if err != nil {
			return "",err
		}

		defer resp.Body.Close()
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			return "",err
		}
		return string(body), nil
    }else{
		if data_enc != "" {
			url+="?"+data_enc
		}
		resp, err := http.Get(url)
		if err != nil {
			return "",err
		}

		defer resp.Body.Close()
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			return "",err
		}
		return string(body), nil
	}
}*/

func make_HTTP_request(queryurl string, data_enc string, post bool) (string) {
	reqstr:=""
	if post{
		reqstr="POST "+queryurl+" HTTP/1.1\r\n"
	}else{
		reqstr="GET "+queryurl+"?"+data_enc+" HTTP/1.1\r\n"
	}
	reqstr+=
		"Host: "+host+"\r\n"+
		"User-Agent: Team7-Golang-TCP"+"\r\n"+
		"Connection: keep-alive"+"\r\n"+
		"Keep-Alive: max=10000, timeout=120"+"\r\n"
		//"Connection: close"+"\r\n"
		//conn close for HTTP once
	if post{
		reqstr+=
		"Content-Type: application/x-www-form-urlencoded"+"\r\n"+
		"Content-Length: "+strconv.Itoa(len(data_enc))+"\r\n"+
		"\r\n"+
		data_enc+"\r\n"
	}
	reqstr+="\r\n" //double line-break for end-of-request
	return reqstr
}

func make_fastconn() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", host, time.Second*5)
	conn.SetDeadline(time.Now().Add(time.Second*5))
	return conn, err
}
func make_longconn() (net.Conn, error) {
	conn, err := net.Dial("tcp", host)
	return conn, err
}

func tcp_HTTP_once(queryurl string, data_enc string, post bool) (string, error) {
	conn, err := make_fastconn()
	if err != nil {return "", err}
	reqstr:=make_HTTP_request(queryurl, data_enc, post)
	_, err = conn.Write([]byte(reqstr))
    if err != nil {return "", err}
		const bufsize=1024
	ret:=""
	for {
		reply := make([]byte, bufsize)
		cnt, err := conn.Read(reply)
		println("read:",string(cnt))
		if err!=nil{
			return ret,err
		}
		ret+= string(reply[:cnt])
		if(cnt<bufsize){
			break
		}
	}
	index:=strings.Index(ret,"\r\n\r\n")
	defer conn.Close()
	return ret[index+4:],nil
}

func tcp_HTTP_multirep(queryurl string, data_enc []string, post bool){
	conn, _ := make_longconn()
	//if err != nil {return "", err}
	reqstr:=""
	for i:=0;i<len(data_enc);i++ {
		reqstr+=make_HTTP_request(queryurl, data_enc[i], post)
	}
	println("sending...:",reqstr)
	_, err := conn.Write([]byte(reqstr))
	if err!=nil{
		println("err",err,err.Error())
		return
	}
	const bufsize=1024
	ret:=""
	words:= regexp.MustCompile("\\r\\n\\r\\n")
    for {
		reply := make([]byte, bufsize)
		cnt, err := conn.Read(reply)
		println("read:",string(cnt))
		if err!=nil && err.Error()!="EOF"{
			println("err",err,err.Error())
		//	return
		}
		ret+= string(reply[:cnt])
		//if(cnt<bufsize){
		rncnt:=len(words.FindAll([]byte(ret), -1))
		println("rncnt:",rncnt)
		if( rncnt == len(data_enc) ) {
			break
		}
		time.Sleep(1000*time.Millisecond)
	}

	println(ret)
}
func tcp_HTTP_multi(queryurl []string, data_enc []string, post []bool){

}

func do_insert(key string, value string, c chan time.Duration){
	start := time.Now()
	//naive_HTTP(kvURL+"insert","key="+key+"&value="+value,true)
	c<- time.Since(start)
}
//func do_remove



type duration_slice []time.Duration
func (a duration_slice) Len() int { return len(a) }
func (a duration_slice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a duration_slice) Less(i, j int) bool { return a[i] < a[j] }

func main(){
  N:=10

  ret,err := tcp_HTTP_once("/kvman/dump","",false)
  if err!=nil{
    fmt.Println(err)
	os.Exit(-1)
  }
  fmt.Println(ret)

  data_enc:=make([]string,N)
  for i:=0; i<N;i++ {
	data_enc[i]="key="+strconv.Itoa(i)+"&value="+strconv.Itoa(i)
  }
  tcp_HTTP_multirep("/kv/insert",data_enc,true)

  /*
  dummy:="TEST keyvalue long string................"
  for i:=0; i<1000;i++ {
	dummy=dummy+ string(i%26+65)
  }

  perf:=make(chan time.Duration, N)

  for i:=0; i<N;i++ {
	go do_insert(dummy+strconv.Itoa(i),strconv.Itoa(i)+dummy, perf)
  }

  stat:=make(duration_slice, N)  // [N]time.Duration
  //stat:=make([]int64, N)
  for i:=0; i<N;i++ {
	stat[i]= <-perf
  }
  sort.Sort(stat)

  for i:=1;i<=9;i++ {
	fmt.Print(strconv.Itoa(i*10)+"% Percentile:")
	fmt.Println(stat[i*N/10])
  }*/
  stat:=make(duration_slice, N)  // [N]time.Duration
  //stat:=make([]int64, N)

  sort.Sort(stat)
}
