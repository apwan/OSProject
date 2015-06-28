package kvpaxos

import (
  "net"
  "net/http"
  "net/rpc"
  "sync"
  "os"
  "syscall"
  "encoding/gob"
  "encoding/json"
  "math/rand"
  "time"
  "strconv"
  "fmt"
  "log"

  "paxos"
  "kvlib"
  "stoppableHTTPlistener"
  )

const SaveMemThreshold=10
  const Debug=false
const StartHTTP=true

func DPrintf(format string, a ...interface{}) (n int, err error) {
  if Debug {
    log.Printf(format, a...)
  }
  return
}


type Op struct {
  // Your definitions here.
  // Field names must start with capital letters,
  // otherwise RPC will break.
  IsPut bool // type
  Key string
  Value string
  Who int
  OpID int
}

func DeepCompareOps(a Op, b Op) (bool){
  return a.IsPut==b.IsPut &&
  a.Key==b.Key &&
  a.Value==b.Value &&
  a.OpID==b.OpID &&
  a.Who==b.Who
}

type KVPaxos struct {
  mu sync.Mutex
  l net.Listener
  me int
  N int
  dead bool // for testing
  unreliable bool // for testing
  px *paxos.Paxos

  px_touchedPTR int

  snapshot map[string]string
  snapstart int

  HTTPListener *stoppableHTTPlistener.StoppableListener
}


func (kv *KVPaxos) PaxosStatOp() (int,map[string]string) {
    if Debug{
        fmt.Printf("Paxos STAT!\n")
    }
    kv.mu.Lock(); // Protect px.instances
    defer kv.mu.Unlock();
    
    //need to insert a meaningless OP, in order to sync DB!
    var myop Op
    myop.IsPut=false
    myop.Key=""
    myop.OpID=rand.Int()
    myop.Who=-1
    var ID int
    var value interface{}
    var decided bool

    //check if there's existing same OP...
    {
      ID=kv.px_touchedPTR+1
      //ID=0
      for true {
          kv.px.Start(ID,myop)
          time.Sleep(50)
          for true {
              decided,value = kv.px.Status(ID)
              if decided {
                  break;
              }
              time.Sleep(200*time.Millisecond)
          }
          //if DeepCompareOps(value.(Op),myop) {//succeeded
          //    if Debug {fmt.Printf("Saw DCSame! %v %v server%d\n",value,myop,kv.me)}
          //    break;
          //}
          if value.(Op).OpID==myop.OpID {//succeeded
          //    if Debug {fmt.Printf("Saw OIDSame but DC fail! %v %v\n",value,myop)}
              break;
          }
          var scale=(kv.me+ID)%3
          time.Sleep(time.Duration(rand.Intn(10)*scale*int(time.Millisecond)))
          ID++
      }
      kv.px_touchedPTR=ID
    }



    tmp:=make(map[string]string)
    for k,v:=range kv.snapshot {
      tmp[k]=v
    }
    for i:=kv.snapstart;i<=kv.px_touchedPTR;i++{
        _,value := kv.px.Status(i)
        v:=value.(Op)
        tmp[v.Key]=v.Value
    }
    var cnt=0
    tmp2:=make(map[string]string)
    for k,v:=range tmp {
      if v!=""{
       tmp2[k]=v
       cnt++
      }
    }
    return cnt,tmp2
}

func (kv *KVPaxos) PaxosAgreementOp(isput bool, opkey string, opvalue string, who int, opid int) (Err,string) {//return (Err,value)
    if Debug{
        fmt.Printf("P/G Step0, isput:%d\n",isput)
    }
    kv.mu.Lock(); // Protect px.instances
    defer kv.mu.Unlock();
    if Debug {
        println("P/G Step1")
    }

       //step1: get the agreement!
    var myop Op
    myop.IsPut=isput
    myop.Key=opkey
    myop.OpID=opid
    if isput{
      myop.Value=opvalue
    }
    myop.Who=who
    var ID int
    var value interface{}
    var decided bool

    //check if there's existing same OP...
    var sameID=-1
    for i:=kv.snapstart;i<=kv.px_touchedPTR;i++{
      decided,value = kv.px.Status(i)
      if decided {
        //if DeepCompareOps(value.(Op),myop){
        if value.(Op).OpID==myop.OpID{
          sameID=i
          if Debug {fmt.Printf("Saw sameID! id%d opid%d sv#%d",sameID,myop.OpID,kv.me)}
          break
        }
      }else {
        fmt.Printf("PANIC %v %v\n", value, myop);
        panic("Not decided, but before touchPTR??")
      }
    }
    if sameID>=0{
      ID=sameID//skip
    }else{
      ID=kv.px_touchedPTR+1
      //ID=0
      for true {
          kv.px.Start(ID,myop)
          time.Sleep(50)
          for true {
              decided,value = kv.px.Status(ID)
              if decided {
                  break;
              }
              time.Sleep(200*time.Millisecond)
          }
          if DeepCompareOps(value.(Op),myop) {//succeeded
              if Debug {fmt.Printf("Saw DCSame! %v %v server%d\n",value,myop,kv.me)}
              break;
          }
          if value.(Op).OpID==myop.OpID {//succeeded
              if Debug {fmt.Printf("Saw OIDSame but DC fail! %v %v\n",value,myop)}
              break;
          }
          var scale=(kv.me+ID)%3
          time.Sleep(time.Duration(rand.Intn(10)*scale*int(time.Millisecond)))
          ID++
      }
      kv.px_touchedPTR=ID
    }

    if Debug {fmt.Printf("Decided! %d=%v server%d\n",ID,myop,kv.me)}

    if Debug {
        println("P/G Step2")
    }
    //We got ID!
    //Step2: trace back for previous value
    var latestVal string
    var latestValFound=false
    for i:=ID-1;i>=kv.snapstart;i--{ //i>=latest snapshot!
        de,op:=kv.px.Status(i)
        for de==false {
            time.Sleep(10*time.Millisecond)
            de,op=kv.px.Status(i)
        }
        optt,found:=op.(Op)
        if found==false{
            return "Not Found type .(Op)",""
        }
        if optt.IsPut==false{
            continue;
        }
        //if optt.OpID==myop.OpID{
         //   fmt.Printf("Saw same OpID %d who %d %d val %s %s \n",myop.OpID,myop.Who,optt.Who,myop.Value,optt.Value);
        //    fmt.Printf("Saw key %s %s \n",myop.Key,optt.Key);
        //    continue;
        //}
        if optt.Key==myop.Key{
          if optt.OpID==myop.OpID{continue;}
            latestVal=optt.Value
            latestValFound=true
            break
        }
    }
    //new step2.5: check snapshot!
    if latestValFound==false {
        v, ok := kv.snapshot[myop.Key];
        if ok {
          latestValFound=true
          latestVal=v
        }
    }


    //returning value...
    if isput{
      return "",latestVal
    }else{
        if latestValFound==false {//equivalently, i<0
            return "Key Not Found",""
        }
        return "",latestVal
    }

}

func (kv *KVPaxos) Get(args *GetArgs, reply *GetReply) error {
  e,Value:=kv.PaxosAgreementOp(false,args.Key,"",args.ClientID,args.OpID)
  reply.Err=e
  reply.Value=Value
  return nil

}

func (kv *KVPaxos) Put(args *PutArgs, reply *PutReply) error {
  e,Value:=kv.PaxosAgreementOp(true,args.Key,args.Value,args.ClientID,args.OpID)
  reply.Err=e
  reply.PreviousValue=Value
  return nil
}

// tell the server to shut itself down.
// please do not change this function.
func (kv *KVPaxos) kill() {
  DPrintf("Kill(%d): die\n", kv.me)
  kv.dead = true
  kv.l.Close()
  kv.px.Kill()
  if StartHTTP{
    println("Stopping HTTP...")
    kv.HTTPListener.Stop()
  }
}
func (kv *KVPaxos) Kill() {//public wrapper
  kv.kill()
}

func (kv *KVPaxos) DumpInfo() string {
  r:=""
  r+=fmt.Sprintf("I'm %d\n",kv.me)
  r+=fmt.Sprintf("Max pxID=%d\n",kv.px.Max())
  r+=fmt.Sprintf("Min pxID=%d\n",kv.px.Min())
  r+=fmt.Sprintf("PTR pxID=%d\n",kv.px_touchedPTR)

  ID:=kv.px.Max()
  for i:=0;i<=ID;i++ {
    de,op:=kv.px.Status(i)
    o,_:=op.(Op)
    if de {
      tmp := "GET"
      if o.IsPut {
        tmp = "PUT"
      }
      r+=fmt.Sprintf("Op[%d] %s %s=%s by%d  opid%d\n",i,tmp,o.Key,o.Value,o.Who,o.OpID)
    }else{
      r+=fmt.Sprintf("Op[%d] undecided  \n",i)
    }
  }
  //r+="<meta http-equiv=\"refresh\" content=\"1\">"
  return r
}


func (kv *KVPaxos) housekeeper() {
  for true{
    if kv.dead {
      if Debug{println("KVDB dead, housekeeper done") }
      break
    }
    time.Sleep(time.Second*2)
    curr:=kv.px.Max()-1
    mem:=kv.snapstart
    if Debug {fmt.Printf("hosekeeper #%d, max %d, snap %d... \n",kv.me,curr,mem) }
    if(curr-mem> SaveMemThreshold){//start compressing...
      println("Housekeeper GC starting...");
      kv.mu.Lock(); // Protect px.instances
        curr-=5
        for i:=kv.snapstart;i<curr;i++ {
          de,op:=kv.px.Status(i)
          if de==false {
            break
          }
          optt,found:=op.(Op)
          if found==false{
              println("Housekeeper error! Not Found type .(Op)")
              break
          }
          if optt.IsPut==true{
              kv.snapshot[optt.Key]=optt.Value
          }
          kv.px.Done(i)
          kv.snapstart=i+1
        }
      kv.mu.Unlock();
      if Debug {fmt.Printf("done!#%d now: max %d, min %d, snap %d...\n",kv.me,kv.px.Max(),kv.px.Min(),kv.snapstart) }
    }
  }
}

//HTTP handlers generator; to create a closure for kvpaxos instance
func kvDumpHandlerGC(kv *KVPaxos) http.HandlerFunc{
  return func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "%s",kv.DumpInfo())
  }
}
var globalOpsCnt=0
func kvPutHandlerGC(kv *KVPaxos) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    key:= r.FormValue("key")
    value:= r.FormValue("value")
    opid:= r.FormValue("id")
    if value=="" {
      fmt.Fprintf(w, "%s",kvlib.JsonErr("value not found, please give nonempty string"))
      return
    }


    var args PutArgs = PutArgs{key,value,true,globalOpsCnt+kv.me,-1}
    globalOpsCnt+=kv.N
    var reply PutReply = PutReply{"",""}
    if opid!="" {
      args.OpID,_=strconv.Atoi(opid)
    }

    err:=kv.Put(&args,&reply)
    if err!=nil || reply.Err!=""{
      fmt.Fprintf(w, "%s",kvlib.JsonErr(string(reply.Err)))
      return
    }
    fmt.Fprintf(w, "%s",kvlib.JsonSucc(reply.PreviousValue))
  }
}

func kvGetHandlerGC(kv *KVPaxos) http.HandlerFunc{
  return func(w http.ResponseWriter, r *http.Request) {
    key:= r.FormValue("key")
    opid:= r.FormValue("id")

    var args GetArgs = GetArgs{key,globalOpsCnt+kv.me,-1}
    globalOpsCnt+=kv.N
    var reply GetReply = GetReply{"",""}
    if opid!="" {
      args.OpID,_=strconv.Atoi(opid)
    }

    err:=kv.Get(&args,&reply)
    if err!=nil || reply.Err!=""{
      fmt.Fprintf(w, "%s",kvlib.JsonErr(string(reply.Err)))
      return
    }
    fmt.Fprintf(w, "%s",kvlib.JsonSucc(reply.Value))
  }
}
func kvGeneralHandlerGC(kv *KVPaxos) http.HandlerFunc{
  return func(w http.ResponseWriter, r *http.Request) {
    key:= r.FormValue("key")
    value:= r.FormValue("value")
    opid:= r.FormValue("id")
    simulate_method:= r.FormValue("method")

    if key=="" {
      fmt.Fprintf(w, "Usage: Send GET/PUT/DELETE/UPDATE requests to /kv/?key=..&value=..")
      return
    }
    method := r.Method
    if method=="GET" && simulate_method!=""{
      method=simulate_method
    }

    var ID=globalOpsCnt+kv.me
    globalOpsCnt+=kv.N
    if opid!="" {
      ID,_=strconv.Atoi(opid)
    }

    //var retVal=""
    //var succ=false

    switch method {
      case "GET":
        var reply GetReply = GetReply{"",""}
        var args GetArgs = GetArgs{key,ID,-1}
        err:=kv.Get(&args,&reply)
        if err!=nil || reply.Err!=""{
          fmt.Fprintf(w, "%s",kvlib.JsonErr(string(reply.Err)))
        }else{
          fmt.Fprintf(w, "%s",kvlib.JsonSucc(reply.Value))
        }
        return
      case "UPDATE":
      case "PUT":
      case "DELETE":
      default:
        fmt.Fprintf(w, "Usage: Send GET/PUT/DELETE/UPDATE requests to /kv/?key=..&value=..")
        return
    }

    if method=="DELETE" {
      value=""
    }

    //!! The UPDATE is now incorrect!!

    var args PutArgs = PutArgs{key,value,true,ID,-1}
    var reply PutReply = PutReply{"",""}
    
    err:=kv.Put(&args,&reply)

    if err!=nil || reply.Err!=""{
      fmt.Fprintf(w, "%s",kvlib.JsonErr(string(reply.Err)))
      return
    }
    fmt.Fprintf(w, "%s",kvlib.JsonSucc(reply.PreviousValue))

  }
}


/*
func kvmanCountkeyHandler(w http.ResponseWriter, r *http.Request) {
  if check_HTTP_method && r.Method != "GET" {
    fmt.Fprintf(w, "Bad Method: Please use GET")
    return
  }
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
  if check_HTTP_method && r.Method != "GET" {
    fmt.Fprintf(w, "Bad Method: Please use GET")
    return
  }
  var str,err=db.MarshalJSON();
  if err==nil{
    fmt.Fprintf(w, "%s",str)
    return
  }
  fmt.Fprintf(w, "DB marshalling error %s",err)
}
func kvmanShutdownHandler(w http.ResponseWriter, r *http.Request) {*/


func kvmanCountKeyHandlerGC(kv *KVPaxos) http.HandlerFunc{
  return func(w http.ResponseWriter, r *http.Request) {
    cnt,_:=kv.PaxosStatOp()
    tmp := make(map[string]int)
    tmp["result"]=cnt
    var str,_=json.Marshal(tmp)
    fmt.Fprintf(w, "%s",str)
  }
}
func kvmanDumpHandlerGC(kv *KVPaxos) http.HandlerFunc{
  return func(w http.ResponseWriter, r *http.Request) {
    _,data:=kv.PaxosStatOp()
    var str,_=json.Marshal(data)
    fmt.Fprintf(w, "%s",str)
  }
}
func kvmanShutdownHandlerGC(kv *KVPaxos) http.HandlerFunc{
  return func(w http.ResponseWriter, r *http.Request) {
    defer func(){
      time.Sleep(1)
      kv.Kill()
    }()
    fmt.Fprintf(w, "{success:\"true\",message:\"The kvpaxos server will shutdown immediately. Please wait for 10ms before the HTTP server detach (and release the listening port).\"}")
  }
}
//end HTTP handlers

var RPC_Use_TCP int = 0

//
// servers[] contains the ports of the set of
// servers that will cooperate via Paxos to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
//
func StartServer(servers []string, me int) *KVPaxos {
  if RPC_Use_TCP == 1{
    paxos.RPC_Use_TCP = 1
  }
  // call gob.Register on structures you want
  // Go's RPC library to marshall/unmarshall.
  gob.Register(Op{})

  kv := new(KVPaxos)
  kv.me = me
  kv.N = len(servers) //used for universal incrementation of HTTP request OpIDs
  kv.px_touchedPTR=-1 //0 is untouched at the beginning!
  kv.snapstart=0
  kv.snapshot=make(map[string]string)

  go kv.housekeeper()
  // Your initialization code here.

  if StartHTTP{

    //HTTP initialization
    //wait for a while, since previous server hasn't timed out on TCP!
    time.Sleep(time.Millisecond*11)

    serveMux := http.NewServeMux()

    var kvHandlerGCs = map[string]func(*KVPaxos)http.HandlerFunc{
      "dump": kvDumpHandlerGC,
      "put": kvPutHandlerGC,
      "get": kvGetHandlerGC,
    }
    var kvmanHandlerGCs = map[string]func(*KVPaxos)http.HandlerFunc{
      "countkey": kvmanCountKeyHandlerGC,
      "dump": kvmanDumpHandlerGC,
      "shutdown": kvmanShutdownHandlerGC,
    }

    for key,val := range kvHandlerGCs{
      serveMux.HandleFunc("/"+key, val(kv))
    }
    for key,val := range kvmanHandlerGCs{
      serveMux.HandleFunc("/kvman/"+key, val(kv))
    }

    serveMux.HandleFunc("/kv/", kvGeneralHandlerGC(kv))

    confname := "conf/settings.conf"

    if _,err:=os.Stat(confname); err!=nil && os.IsNotExist(err){
      confname = "../../" + confname;
    }

    conf:= kvlib.ReadJson(confname)
    listenPort:=kvlib.Find_Port(me,conf)
    s := &http.Server{
      //Addr: ":"+strconv.Itoa(listenPort),
      Handler: serveMux,
      ReadTimeout: 1 * time.Second,
      WriteTimeout: 30 * time.Second,
      MaxHeaderBytes: 1<<20,
    }

    originalListener, err := net.Listen("tcp", ":"+strconv.Itoa(listenPort))
    if err!=nil {
      panic(err)
    }
    sl, err := stoppableHTTPlistener.New(originalListener)
    if err!=nil {
      panic(err)
    }
    kv.HTTPListener=sl
    go func(){
      fmt.Printf("Starting HTTP server: %d\n",listenPort)
      s.Serve(sl)
      //will be stopped by housekeeper!
    }()

  }



  // End of initialization code

  rpcs := rpc.NewServer()
  rpcs.Register(kv)

  kv.px = paxos.Make(servers, me, rpcs)

  os.Remove(servers[me])
  l, e := net.Listen("unix", servers[me]);
  if e != nil {
    log.Fatal("listen error: ", e);
  }
  kv.l = l


  // please do not change any of the following code,
  // or do anything to subvert it.

  go func() {
    for kv.dead == false {
      conn, err := kv.l.Accept()
      if err == nil && kv.dead == false {
        if kv.unreliable && (rand.Int63() % 1000) < 100 {
          // discard the request.
          conn.Close()
        } else if kv.unreliable && (rand.Int63() % 1000) < 200 {
          // process the request but force discard of reply.
          c1 := conn.(*net.UnixConn)
          f, _ := c1.File()
          err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
          if err != nil {
            fmt.Printf("shutdown: %v\n", err)
          }
          go rpcs.ServeConn(conn)
        } else {
          go rpcs.ServeConn(conn)
        }
      } else if err == nil {
        conn.Close()
      }
      if err != nil && kv.dead == false {
        fmt.Printf("KVPaxos(%v) accept: %v\n", me, err.Error())
        kv.kill()
      }
    }
  }()

  return kv
}
