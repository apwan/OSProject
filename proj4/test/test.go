package main

import(
  "net/http"
  "fmt"
  "os"
  "time"
  "log"
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
)


func start_server_Handler(w http.ResponseWriter, r *http.Request) {
	res := StartServer(strconv.Itoa(role))
	fmt.Fprintf(w, "Start Server %d: %s",role, res)
}
func stop_server_Handler(w http.ResponseWriter, r *http.Request) {
	res := StopServer(strconv.Itoa(role))
	fmt.Fprintf(w, "Stop Server %d: %s",role, res)
}
func shutdown_Handler(w http.ResponseWriter, r *http.Request){
  fmt.Fprintf(w, "Goodbye main tester! Tester %d shutdown!", role)
  fmt.Printf("Tester %d shutdown!\n", role)
  go func(){
    time.Sleep(time.Millisecond*500) //sleep epsilon
    os.Exit(role)
  }()

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
    time.Sleep(time.Millisecond * 5000)

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
    http.HandleFunc("/test/stop_server", stop_server_Handler)
    http.HandleFunc("/test/shutdown", shutdown_Handler)
    log.Fatal(s.ListenAndServe())
  }
}
