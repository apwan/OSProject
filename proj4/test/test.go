package main

import(
  "net/http"
  "fmt"
  "os"
  "os/exec"
  "time"
  "log"
  "strings"
  "strconv"
  //our lib
  . "kvlib"

)

func usage(){
  fmt.Println("The main tester calls other remote testers to start/stop server.")
  fmt.Println("[id]   :    Launch the remote tester of specified id.")
  fmt.Println("-m     :    Launch the main tester.")
  os.Exit(1)
}


var(
  role = Det_role()
  pr *os.Process = nil // keep the kv server process for killing
  remaining = 0
  not_forced_to_quit = true
)

func check_alive_Handler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "I am alive!")
  return
}

func start_server_Handler(w http.ResponseWriter, r *http.Request) {
  if pr != nil{
    fmt.Fprintf(w, "Server %d Already Started: %s",role, pr)
    return
  }
  attr := &os.ProcAttr{
        Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
    }
    id := fmt.Sprintf("n%02d", role)
  pr,_ = os.StartProcess("bin/start_server", []string{"bin/starst_server",id}, attr)
	//res := CmdStartServer(strconv.Itoa(role))
	fmt.Fprintf(w, "Start Server %d: %s",role, pr)
}
func kill_server_Handler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("kill handler")
  if pr==nil{
    fmt.Fprintf(w, "No server %d process to kill", role)
    return
  }
  err := pr.Kill()
  if err!=nil{
    fmt.Fprintf(w, "Kill Server %d Failed: %s", role, err)
    pr.Kill()
  }else{
    fmt.Fprintf(w, "Kill Server %d Success!", role)
  }
  pr = nil
}
func stop_server_Handler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("stop handler")
  var res string
  id := fmt.Sprintf("n%02d", role)
  cmd := exec.Command("bin/stop_server", []string{id}...)
  o,e := cmd.Output()

  if e!=nil || len(o)<=1 {
    res = "Error"
  }else{
    res = strings.Replace(string(o[:]),"\n"," ",-1)
  }
  pr = nil
	fmt.Fprintf(w, "Tester %d Stop Server: %s",role, res)
}
func shutdown_Handler(w http.ResponseWriter, r *http.Request){
  fmt.Fprintf(w, "Goodbye main tester! Tester %d shutdown!", role)
  if pr!=nil{
    pr.Kill()
  }
  fmt.Printf("Tester %d shutdown!\n", role)
  go func(){
    time.Sleep(time.Millisecond*1000) //sleep epsilon
    os.Exit(role)
  }()
  return
}

func main_Handler(w http.ResponseWriter, r *http.Request){
  key:= r.FormValue("op")
  switch key{
  case "finish":
    val := r.FormValue("forced")
    if val=="true"{
      fmt.Fprintf(w,"Forced to finish!")
      go func(){
        All_request(Server_addr, "/kvman/shutdown")
        All_request(Tester_addr, "/test/shutdown")
        time.Sleep(time.Millisecond * 1000)
        os.Exit(0)
      }()
      return
    }else if val == "report"{
      not_forced_to_quit = false
      fmt.Fprintf(w, "Remain: %d. Report after current test case!", remaining)
      return
    }else{
      pre_remain := remaining
      count := 30
      flag := 0
      fmt.Fprintf(w, "Remain: %d. Will finish after current test case!", pre_remain)
      go func(){
        for pre_remain<=remaining && count>0{
          time.Sleep(time.Millisecond * 5000)
          fmt.Println(count,"\tWait for current case to finish!\n")
          count--
        }
        All_request(Server_addr, "/kvman/shutdown")
        All_request(Tester_addr, "/test/shutdown")
        flag = 1
      }()

      go func(){
        for flag==0 {
          time.Sleep(time.Millisecond * 10000)
        }
        os.Exit(0)
      }()

    }

  case "checkalive":
    fmt.Fprintf(w, "I am alive!")
    return
  default:
    fmt.Fprintf(w, "Usage: /main?op=finish&forced=true/false")
    return
  }
}

// run on main tester
func StartTest(conf map[string]string) string{
  res := ""
  // may need to check using the same port

  res += fmt.Sprintf("config: \n\t srv01: %s\n\t srv02: %s\n\t srv03: %s\n",
    Server_addr[0],Server_addr[1],Server_addr[2])

  tot,_ := strconv.Atoi(conf["test_total"])
  cnt := tot
  remaining = tot
  fmt.Println("********************* Start Testing *****************************")
  var dura_total time.Duration
  for i := 0; i < tot && not_forced_to_quit ; i++ {
    testname := conf["pre"]+strconv.Itoa(i)+".test"
    if conf["fmt"] != "true"{ //no need to specify each test case name
      testname = conf["pre"]+conf[strconv.Itoa(i)]
    }
    auto_restart := false
    if conf["auto_restart_server"]=="true"{
      All_request(Tester_addr, "/test/start_server")
      auto_restart = true
      time.Sleep(time.Millisecond * 1500)
    }



      start_time := time.Now()
      res, fail := TestUnit(Server_addr, Tester_addr, testname, auto_restart)
      end_time :=time.Now()
      var dura time.Duration = end_time.Sub(start_time)
      dura_total += dura
      if conf["auto_restart_server"]=="true"{
        All_request(Tester_addr, "/test/stop_server")
        time.Sleep(time.Millisecond * 500)
      }
      if conf["with_err_msg"]=="true"{
        fmt.Printf("%s", res)
        if fail == 0 {
            fmt.Printf("\nTest case %d: success!\n", i)
        } else {
            fmt.Printf("\nTest case %d: failed!\n", i)
        }
        fmt.Printf("Ellapsed Time: %f secs\n\n", dura.Seconds())
      }

      cnt -= fail
      remaining--

  }
  fmt.Println("********************* Finish Testing *****************************")

  if cnt == tot {
    res+="\nSuccess\n"
  } else {
    res+="\nFail\n"
  }
  res+= fmt.Sprintf("Total Time: %f mins\n\n", dura_total.Minutes())

  fmt.Println(res)

  return res
}


func All_request(addr []string, req string){
  for i:=0;i<len(addr);i++ {
    resp, err := http.Get(addr[i] +req)
    if err != nil{
      fmt.Println(err)
    }else{
      fmt.Println(DecodeStr(resp))
    }
  }
}

var auxTesterHandlers map[string]http.HandlerFunc = map[string]http.HandlerFunc{
    "/test":check_alive_Handler,
    "/test/start_server":start_server_Handler,
    "/test/stop_server":kill_server_Handler,
    "/test/shutdown":shutdown_Handler,
}

var Tester_addr []string = make([]string, 3)
var Server_addr []string = make([]string, 3)

func mainTesterCheck(addr []string)bool{
    count := 0
    for i:=0;i<len(addr);i++{
      _,err := http.Get(addr[i]+"/test")
      if err != nil{
        fmt.Println(err)
        fmt.Printf("Remote tester %d not alive!\n", i+1)
      }else{
        count ++
      }
    }
    return count >= len(addr)
}


func main(){
  if role < 0{
    fmt.Printf("role: %d\n", role)
    usage()
  }
  if role == 0{
    fmt.Println("Launch main tester")
  }else{
    fmt.Printf("Launch tester %d\n",role)
  }
  conf := ReadJson("conf/test.conf")
  ips := []string{"127.0.0.1",conf["n01"],conf["n02"],conf["n03"]}
  ports := []string{conf["port"],conf["port_n01"],conf["port_n02"],conf["port_n03"]}
  if conf["use_different_port"]!="true"{
    ports[1],ports[2],ports[3] = ports[0],ports[0],ports[0]
  }
  tester_ports := []string{conf["test_port00"],conf["test_port01"],conf["test_port02"],conf["test_port03"]}

  for i:=1;i<=3;i++{
    Tester_addr[i-1] = fmt.Sprintf("http://%s:%s" ,ips[i],tester_ports[i])
    Server_addr[i-1] = fmt.Sprintf("http://%s:%s" ,ips[i], ports[i])
  }

  if role == 0{
    // check connections with remote testers
    for i:=0;i<30;i++{
      if mainTesterCheck(Tester_addr){
        fmt.Println("All remote testers alive!")
        fmt.Println("Main tester start testing!")
        break
      }else{
        fmt.Println("Main tester wait for another check")
        time.Sleep(time.Millisecond * 2000)
      }
    }
    defer All_request(Tester_addr, "/test/shutdown") //remember to shutdown remote testers

    s := &http.Server{
  		Addr: ":"+tester_ports[role],
  		Handler: nil,
  		ReadTimeout: 10 * time.Second,
  		WriteTimeout: 10 * time.Second,
  		MaxHeaderBytes: 1<<20,
  	}
    http.HandleFunc("/main", main_Handler)
    go func(){
      log.Fatal(s.ListenAndServe())
    }()

    /* run test cases here !  */


    StartTest(conf)


    /* all test cases finished ! */

    time.Sleep(time.Millisecond * 2000)

    fmt.Println("Main tester finished!\n")

  }else{
    s := &http.Server{
  		Addr: ":"+tester_ports[role],
  		Handler: nil,
  		ReadTimeout: 10 * time.Second,
  		WriteTimeout: 10 * time.Second,
  		MaxHeaderBytes: 1<<20,
  	}
    if conf["stop_by_kill"]!="true"{
      auxTesterHandlers["/test/stop_server"] = stop_server_Handler
    }

    for key,val := range auxTesterHandlers{
      http.HandleFunc(key,val)
    }

    log.Fatal(s.ListenAndServe())
  }
}
