package main

import(
  "net/http"
  "fmt"
  "os"
  "os/exec"
  "time"
  "log"
  "bytes"
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
  pr *os.Process = nil
)


func start_server_Handler(w http.ResponseWriter, r *http.Request) {
  if pr != nil{
    fmt.Fprintf(w, "Server %d Already Started: %s",role, pr)
    return
  }
  attr := &os.ProcAttr{
        Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
    }

  pr,_ = os.StartProcess("bin/start_server", []string{"bin/starst_server",fmt.Sprintf("n%02d", role)}, attr)
	//res := CmdStartServer(strconv.Itoa(role))
	fmt.Fprintf(w, "Start Server %d: %s",role, pr)
}
func kill_server_Handler(w http.ResponseWriter, r *http.Request) {
  err := pr.Kill()
	//res := CmdStopServer(strconv.Itoa(role))
	fmt.Fprintf(w, "Stop Server %d: %s",role, err)
}
func stop_server_Handler(w http.ResponseWriter, r *http.Request) {
  if pr==nil {
    fmt.Fprintf(w, "Server %d not started yet",role)
    return
  }
  cmd := exec.Command("bin/stop_server", []string{fmt.Sprintf("n%02d", role)}...)
  o,e := cmd.Output()
  var res string
  if e!=nil || len(o)<=1 {
    res = "Error"
  }else{
    res = string(o[:bytes.IndexByte(o,'\n')])
    pr = nil
  }
	fmt.Fprintf(w, "Stop Server %d: %s",role, res)
}
func shutdown_Handler(w http.ResponseWriter, r *http.Request){
  fmt.Fprintf(w, "Goodbye main tester! Tester %d shutdown!", role)
  if pr!=nil{
    pr.Kill()
  }
  fmt.Printf("Tester %d shutdown!\n", role)
  go func(){
    time.Sleep(time.Millisecond*500) //sleep epsilon
    os.Exit(role)
  }()

}

// run on main tester
func StartTest(conf map[string]string) string{
  res := ""
  ips := []string{"127.0.0.1",conf["n01"],conf["n02"],conf["n03"]}
  ports := []string{conf["port"],conf["port_n01"],conf["port_n02"],conf["port_n03"]}
  tester_ports := []string{conf["test_port00"],conf["test_port01"],conf["test_port02"],conf["test_port03"]}
  // may need to check using the same port

  res += fmt.Sprintf("config: \n\t srv01:%s\n\t srv02:%s\n\t srv03:%s\n",ips[1],ips[2],ips[3])
  var addr_pre [3]string
  var tester_addr [3]string
  for i:=0;i<3;i++{
    addr_pre[i] = fmt.Sprintf("http://%s:%s",ips[i+1],ports[i+1])
    tester_addr[i] = fmt.Sprintf("http://%s:%s", ips[i+1], tester_ports[i+1])
  }
  tot,_ := strconv.Atoi(conf["test_total"])
  cnt := tot

  for i := 0; i < tot; i++ {
    testname := conf["pre"]+strconv.Itoa(i)+".test"
    if conf["fmt"] != "true"{ //no need to specify each test case name
      testname = conf["pre"]+conf[strconv.Itoa(i)]
    }
      res, fail := TestUnit(addr_pre, tester_addr, testname)
      if conf["with_err_msg"]=="true"{
        fmt.Printf("%s", res)
        if fail == 0 {
            fmt.Printf("\nTest case %d: success!\n\n", i)
        } else {
            fmt.Printf("\nTest case %d: failed!\n\n", i)
        }


      }
      cnt -= fail

  }

  if cnt == tot {
    res+="Success"
  } else {
    res+="Fail"
  }

  return res
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
  ports := []string{conf["test_port00"],conf["test_port01"],conf["test_port02"],conf["test_port03"]}

  if role == 0{
    // check connections with remote testers
    for i:=1;i<=3;i++ {
      resp, err := http.Get("http://" +ips[i]+ ":" + ports[i] + "/test/start_server")
      if err != nil{
        fmt.Println(err)
      }else{
        fmt.Println(DecodeStr(resp))
      }
    }

    time.Sleep(time.Millisecond * 3000)
    /* run test cases here !  */


    fmt.Println(StartTest(conf))








    /* all test cases finished ! */

    for i:=1;i<=3;i++ {
      resp, err := http.Get("http://" +ips[i]+ ":" + ports[i] + "/test/stop_server")
      if err != nil{
        fmt.Println(err)
      }else{
        fmt.Println(DecodeStr(resp))
      }
    }
    time.Sleep(time.Millisecond * 3000)

    for i:=1;i<=3;i++ {
      resp, err := http.Get("http://" +ips[i]+ ":" + ports[i] + "/test/shutdown")
      if err != nil{
        fmt.Println(err)
      }else{
        fmt.Println(DecodeStr(resp))
      }
    }
    fmt.Println("Main tester finished!\n")

  }else{
    s := &http.Server{
  		Addr: ":"+ports[role],
  		Handler: nil,
  		ReadTimeout: 10 * time.Second,
  		WriteTimeout: 10 * time.Second,
  		MaxHeaderBytes: 1<<20,
  	}
    http.HandleFunc("/test/start_server", start_server_Handler)
    if conf["stop_by_kill"]=="true"{
      http.HandleFunc("/test/stop_server", kill_server_Handler)
    }else{
      http.HandleFunc("/test/stop_server", stop_server_Handler)
    }
    http.HandleFunc("/test/shutdown", shutdown_Handler)
    log.Fatal(s.ListenAndServe())
  }
}
