package main

import(
  "net/http"
  "net/url"
  "fmt"
  "time"
  "strconv"
  "math/rand"
  "sort"
  "os"
  "strings"
  "io/ioutil"
  "encoding/json"
  )


var conf = readConf("conf/settings.conf") 
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
var(
 rootURL="http://"+conf["primary"]+":"+conf["port"]+"/"
 kvURL=rootURL+"kv/"
 kvmanURL=rootURL+"kvman/"
)

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
}
 
var insert_succ=0 
func do_insert(i int, c chan time.Duration){
	key,value:=get_key(i),get_value(i)
	start := time.Now()
	ret,err:=naive_HTTP(kvURL+"insert","key="+url.QueryEscape(key)+"&value="+url.QueryEscape(value),true)
	c<- time.Since(start)
	if err==nil {
		var udat map[string]interface{}
		if err2 := json.Unmarshal([]byte(ret), &udat); err2 == nil {
			if udat["success"]== true {
				insert_succ+=1
			}
		}	
	}
} 
var get_succ=0 
func do_get(i int, c chan time.Duration){
	key,value:=get_key(i),get_value(i)
	start := time.Now()
	ret,err:=naive_HTTP(kvURL+"get","key="+url.QueryEscape(key),true)
	c<- time.Since(start)
	if err==nil {
		var udat map[string]interface{}
		if err := json.Unmarshal([]byte(ret), &udat); err == nil {
			if udat["success"]== true && udat["value"]==value {
				get_succ+=1
			}
		}	
	}
} 
var dummy=func()(string){
  dummy:="TEST keyvalue long string................"
  for i:=0; i<25;i++ {
	dummy=dummy+ string(i%26+65)
  }
  r := rand.New(rand.NewSource(time.Now().UnixNano()))
  dummy=fmt.Sprintf("rand%lf", r.Float64())+dummy+fmt.Sprintf("rand%lf", r.Float64())+":"
  return dummy
}()
func get_key(i int)(string){
	if i%10 ==0  && i<len(dummy){
		copy:=""+dummy
		buf:=[]byte(copy)
		j:=i/10
		buf[j]-=1
		buf[j+1]+=2
		return string(buf)
	}
	if i%2==0{ return strconv.Itoa(i)+dummy }
	return dummy+strconv.Itoa(i)
}
func get_value(i int)(string){
	return strconv.Itoa(i)+dummy+strconv.Itoa(i)
}


type duration_slice []time.Duration
func (a duration_slice) Len() int { return len(a) }
func (a duration_slice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a duration_slice) Less(i, j int) bool { return a[i] < a[j] }

func main(){
  N:=4000
  
  
	arg_num := len(os.Args)
	if arg_num==0 {
		println("Please specify any argument")
		os.Exit(-1)
	}
  
  
  _,err := naive_HTTP(kvmanURL+"dump","",false)
  if err!=nil {
    fmt.Println(err)
	os.Exit(-1)
  }
  //fmt.Println(ret)
  
  insert_perf:=make(chan time.Duration, N)
  for i:=0; i<N;i++ {
	go do_insert(i, insert_perf)
	time.Sleep(time.Microsecond*50)
  }
  insert_stat:=make(duration_slice, N) 
  for i:=0; i<N;i++ {
	insert_stat[i]= <-insert_perf
  }
  sort.Sort(insert_stat)
  
  get_perf:=make(chan time.Duration, N)
  for i:=0; i<N;i++ {
	go do_get(i, get_perf)
	time.Sleep(time.Microsecond*50)
  }
  get_stat:=make(duration_slice, N) 
  for i:=0; i<N;i++ {
	get_stat[i]= <-get_perf
  }
  sort.Sort(get_stat)
  
  time.Sleep(time.Millisecond)
  //println("Insertion: ",insert_succ,"/",N)
  println("Insertion: ",get_succ,"/",N)
  //println("get succ:",get_succ)
  
  var sum_inst=time.Duration(0)
  var sum_get=time.Duration(0)
  for i:=0;i<N;i++ {
	sum_inst+=time.Duration(int(insert_stat[i])/N)
	sum_get+=time.Duration(int(get_stat[i])/N)
  }
  fmt.Print("Average latency: ")
  fmt.Print(sum_inst)
  fmt.Print(" / ")
  fmt.Print(sum_get)
  println()
  
  print("Percentile latency: ")
  
  for i:=2;i<=9;i+=2 {
	//fmt.Print(strconv.Itoa(i*10)+"% Percentile:")
	fmt.Print(insert_stat[i*N/10])
	fmt.Print(" / ")
	fmt.Print(get_stat[i*N/10])
	if(i!=9){print(", ")}
	if(i==2){i++}
  }
  println()
}
