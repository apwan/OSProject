package kvlib

import(
  "net/http"
  "net/url"
  "fmt"
  "time"
  "strconv"
  "sort"
  "strings"
  "encoding/json"

)


func naive_HTTP(url string, data_enc string, post bool) (string, error) {
	if post{
		resp, err := http.Post(url,
			"application/x-www-form-urlencoded",
			strings.NewReader(data_enc))
		if err != nil {
			return "",err
		}
		return DecodeStr(resp), nil
  }else{
		if data_enc != "" {
			url+="?"+data_enc
		}
		resp, err := http.Get(url)
		if err != nil {
			return "",err
		}
		return DecodeStr(resp), nil
	}
}


func Do_insert(i int, kvURL string, c chan time.Duration, insert_succ *int){
	key,value:=get_key(i),get_value(i)
	start := time.Now()
	ret,err:=naive_HTTP(kvURL+"insert","key="+url.QueryEscape(key)+"&value="+url.QueryEscape(value),true)
	c<- time.Since(start)
	if err==nil {
		var udat map[string]interface{}

		if err2 := json.Unmarshal([]byte(ret), &udat); err2 == nil {
      //fmt.Println(udat["success"])
			if udat["success"]== "true" {
				*insert_succ+=1

			}
		}
	}
}

func Do_get(i int, kvURL string, c chan time.Duration, get_succ *int){
	key,value:=get_key(i),get_value(i)
	start := time.Now()
	ret,err:=naive_HTTP(kvURL+"get","key="+url.QueryEscape(key),true)
	c<- time.Since(start)
	if err==nil {
		var udat map[string]interface{}
		if err := json.Unmarshal([]byte(ret), &udat); err == nil {
			if udat["success"]== "true" && udat["value"]==value {
				*get_succ+=1
			}
		}
	}
}
var dummy=GenLongStr()

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

func TestPerformance(N int, kvURL string) string{
  ret := ""
  insert_perf:=make(chan time.Duration, N)
  var insert_succ = 0
  for i:=0; i<N;i++ {
	go Do_insert(i, kvURL, insert_perf, &insert_succ)
  }
  insert_stat:=make(Duration_slice, N)
  for i:=0; i<N;i++ {
	insert_stat[i]= <-insert_perf
  }

  sort.Sort(insert_stat)

  get_perf:=make(chan time.Duration, N)
  var get_succ = 0
  for i:=0; i<N;i++ {
	go Do_get(i, kvURL, get_perf, &get_succ)
  }
  get_stat:=make(Duration_slice, N)
  for i:=0; i<N;i++ {
	get_stat[i]= <-get_perf
  }
  sort.Sort(get_stat)

  time.Sleep(time.Millisecond)
  //println("Insertion: ",insert_succ,"/",N)
  fmt.Println("Insertion: ",get_succ,"/",N)

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

  fmt.Print("\nPercentile latency: ")

  for i:=2;i<=9;i+=2 {
	//fmt.Print(strconv.Itoa(i*10)+"% Percentile:")
	fmt.Print(insert_stat[i*N/10])
  fmt.Print(" / ")
  fmt.Print(get_stat[i*N/10])
	if(i!=9){print(", ")}
	if(i==2){i++}
  }
  fmt.Println()
  return ret
}

// not public

func kvGet(k string, addr string, id int, c chan int, r chan string, start chan bool){
  <-start
  resp, err := http.Get(addr+"/kv/get?key="+k);
  if err == nil {
    r<- DecodeStr(resp)
  }else{
    c<- id
  }
}
func kvInsert(k string, v string, addr string, id int, c chan int, r chan string, start chan bool){
  <- start
  resp, err := http.PostForm(addr + "/kv/insert", url.Values{"key": {k}, "value": {v}})
  if err == nil {
    r<- DecodeStr(resp)
  }else{
    c<- id
  }
}

func kvUpdate(k string, v string, addr string, id int, c chan int, r chan string, start chan bool){
  <- start
  resp, err := http.PostForm(addr + "/kv/update", url.Values{"key": {k}, "value": {v}})
  if err == nil {
    r<- DecodeStr(resp)
  }else{
    c<- id
  }
}

func kvDelete(k string, addr string, id int, c chan int, r chan string, start chan bool){
  <- start
  resp, err := http.PostForm(addr + "/kv/delete", url.Values{"key": {k}})
  if err == nil {
    r<- DecodeStr(resp)
  }else{
    c<- id
  }
}

func testInsert(total_insert int, primary string){
  start := make(chan bool)
  r := make(chan string)
  ch := make(chan int)
  count_fail := 0
  count_success := 0
  for i:=0; i<total_insert; i++ {
    go kvInsert("insert"+strconv.Itoa(i), "res"+strconv.Itoa(i), primary, i, ch, r, start)
  }
  close(start)
  t0 := time.Now()
  for {
    select{
      case f:=<-ch :
        fmt.Println("failed: %d\n", f)
         count_fail ++
      case res:=<-r :
         fmt.Println(res)
         count_success ++
      default:

    }
    if count_fail+count_success>=total_insert{
      break
    }
  }
  t1 := time.Now()
  t := t1.Sub(t0)
  avg := t.Seconds()*1000.0/float64(total_insert)
  fmt.Printf("elapsed %f ms\n", avg)

}

func testGet(total_get int, primary string){
  start := make(chan bool)
  r := make(chan string)
  ch := make(chan int)
  count_fail := 0
  count_success := 0
  for i:=0; i<total_get; i++ {
    go kvGet("insert"+strconv.Itoa(i), primary, i, ch, r, start)
  }
  close(start)
  t0 := time.Now()
  for {
    select{
      case f:=<-ch :
        fmt.Println("failed: %d\n", f)
         count_fail ++
      case <-r :
         //fmt.Println(res)
         count_success ++
      default:

    }
    if count_fail+count_success>=total_get{
      break
    }
  }
  t1 := time.Now()
  t := t1.Sub(t0)
  avg := t.Seconds()*1000.0/float64(total_get)
  fmt.Printf("elapsed %f ms\n", avg )


}
