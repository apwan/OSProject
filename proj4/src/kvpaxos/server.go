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

const (
  PutOp=1
  GetOp=2
  UpdateOp=3
  DeleteOp=4
  NaivePutOp=5
  SaveMemThreshold=15
  Debug=false
  StartHTTP=true
)
var (
  OpName = []string{"NONE","PUT","GET","UPDATE","DELETE","NaivePut"}
)
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
  OpType int // type
  Key string
  Value string
  Who int
  OpID int
}

func DeepCompareOps(a Op, b Op) (bool){
  return a.OpType==b.OpType &&
  a.Key==b.Key &&
  a.Value==b.Value &&
  a.OpID==b.OpID &&
  a.Who==b.Who
}

type ID_Ret_Pair struct {
  OpID int
  Ret string
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

  doneOps map[int]bool
  latestClientOpResult map[int]ID_Ret_Pair
  //Results map[int]string//for debug only

  HTTPListener *stoppableHTTPlistener.StoppableListener
  Death chan int
}


func (kv *KVPaxos) PaxosStatOp() (int,map[string]string) {
    if Debug{
        fmt.Printf("Paxos STAT!\n")
    }
    kv.mu.Lock(); // Protect px.instances
    defer kv.mu.Unlock();

    //need to insert a meaningless OP, in order to sync DB!
    var myop Op = Op{OpType:GetOp, Key:"", Value:"", OpID:rand.Int(),Who:-1}
    var ID int
    var value interface{}
    var decided bool
    //check if there's existing same OP...
    {
      ID=kv.px_touchedPTR+1
      //ID=0
      for !kv.dead {
          kv.px.Start(ID,myop)
          time.Sleep(1)
          for !kv.dead {
              decided,value = kv.px.Status(ID)
              if decided {
                  break;
              }
              time.Sleep(50*time.Millisecond)
          }
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

func (kv *KVPaxos) PaxosAgreementOp(myop Op) (Err,string) {//return (Err,value)
    if Debug{
        fmt.Printf("P/G Step0, OpType:%s\n",OpName[myop.OpType])
    }
    if kv.doneOps[myop.OpID] {
      // Might be the latest op repeated, or an even older one
      lop,found:=kv.latestClientOpResult[myop.Who]
      if found {
        lid:=lop.OpID
        lr:=lop.Ret
        if lid==myop.OpID {
          return "",lr
        }
      }
      return "Error: repeated, old request...","" //should not provide error message, to fall through erroneous ops??
    }

    kv.mu.Lock(); // Protect px.instances
    defer kv.mu.Unlock();
    if Debug {
        println("P/G Step1")
    }

       //step1: get the agreement!

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
      for !kv.dead {
          kv.px.Start(ID,myop)
          time.Sleep(1)
          var backoff time.Duration=10
          for !kv.dead {
              decided,value = kv.px.Status(ID)
              if decided {
                  break;
              }
              time.Sleep(time.Millisecond*backoff)
              if backoff<120{backoff*=2}
          }
          if DeepCompareOps(value.(Op),myop) {//succeeded
              if Debug {fmt.Printf("Saw DCSame! %v %v server%d\n",value,myop,kv.me)}
              break;
          }
          if value.(Op).OpID==myop.OpID {//succeeded
              if Debug {fmt.Printf("Saw OIDSame but DC fail! %v %v\n",value,myop)}
              break;
          }
          var offs uint=uint(ID-kv.px_touchedPTR)
          if offs>4 {offs=4}
          var scale=((kv.me+ID)%3)*(1<<offs)
          time.Sleep(time.Duration(rand.Intn(scale*int(time.Millisecond)+1)))
          ID++
      }
      kv.px_touchedPTR=ID
    }

    if Debug {fmt.Printf("Decided! %d=%v server%d\n",ID,myop,kv.me)}

    if Debug {
        println("P/G Step2")
    }


    var latestVal=kv.snapshot[myop.Key]
    var latestSucc=true
    var beforeVal=""

    var opsVisited=make(map[int]bool)

    var i int
    for i=kv.snapstart;i<=ID;i++{
      decided,value = kv.px.Status(i)
      if !decided {
        fmt.Printf("PANIC %v %v\n", value, myop);
        panic("Not decided, but before touchPTR??")
      }
      var op=value.(Op)
      if Debug{
        fmt.Printf("Simluate Step%d: %d %s %s\n", i, op.OpType,op.Key,op.Value);
      }

      if op.Key!=myop.Key{
        continue
      }


      if opsVisited[op.OpID] {
          continue
      }
      opsVisited[op.OpID]=true
      //do not repeat Ops on unreliable case!

      beforeVal=latestVal
      //this is an op on this key! will update the value!
      switch op.OpType{
        case GetOp:
          latestSucc=(latestVal!="")

        case PutOp:
          if latestVal!=""{
            latestSucc=false
          }else{
            latestVal=op.Value
            latestSucc=true
          }

        case NaivePutOp:
          latestVal=op.Value
          latestSucc=true

        case DeleteOp:
          latestSucc=(latestVal!="")
          latestVal=""

        case UpdateOp:
          if latestVal==""{
            latestSucc=false
          }else{
            latestVal=op.Value
            latestSucc=true
          }
      }
      lm:=len(beforeVal)
      if lm>100{lm=100}
      kv.doneOps[op.OpID]=true;//beforeVal[0:lm]

      l,found:=kv.latestClientOpResult[op.Who]
      if !found || op.OpID>l.OpID { //newer, or not found
        kv.latestClientOpResult[op.Who]=ID_Ret_Pair{op.OpID, beforeVal}
      }
      // should remember the result  if it's the new latest

      if op.OpID==myop.OpID{
        break
      }
         //this result is okay, already
    }

    //i-=1//should be ID, but may not!
    //kv.Results[i]=beforeVal

//    if myop.OpType==GetOp {
//      kv.results[ID]=latestVal
//    }
    //all ops simluated!
    switch myop.OpType{
      case GetOp:
        if latestVal==""{
          return "Key Not Found",""
        }
        return "",latestVal
      case PutOp:
        if !latestSucc{
          return "Put/Insert: key exist?",""
        }
      case DeleteOp:
        if !latestSucc{
          return "Delete: key not exist?",""
        }
      case UpdateOp:
        if !latestSucc{
          return "Update: key not exist?",""
        }
      case NaivePutOp:
    }
    return "",beforeVal

}

func (kv *KVPaxos) Get(args *GetArgs, reply *GetReply) error {
  _,Value:=kv.PaxosAgreementOp(Op{GetOp,args.Key,"",args.ClientID,args.OpID})
  reply.Err=""
  reply.Value=Value
  return nil
}

func (kv *KVPaxos) FormalGet(args *GetArgs, reply *GetReply) error {
  e,Value:=kv.PaxosAgreementOp(Op{GetOp,args.Key,"",args.ClientID,args.OpID})
  reply.Err=e
  reply.Value=Value
  return nil
}

func (kv *KVPaxos) Put(args *PutArgs, reply *PutReply) error {
  //This function is to be called by RPC tester client only. Will always succeed.
  e,Value:=kv.PaxosAgreementOp(Op{NaivePutOp,args.Key,args.Value,args.ClientID,args.OpID})
  reply.Err=e
  reply.PreviousValue=Value
  return nil
}

func (kv *KVPaxos) FormalPut(args *PutArgs, reply *PutReply) error {
  e,Value:=kv.PaxosAgreementOp(Op{PutOp,args.Key,args.Value,args.ClientID,args.OpID})
  reply.Err=e
  reply.PreviousValue=Value
  if Value!=""{
    println(Value)
    panic("Prev Value not empty for formal PUT?")
  }
  return nil
  //will fail if the key exists!
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
    println("HTTP stopped.")
  }
  kv.Death<-1
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
      r+=fmt.Sprintf("Op[%d] %s %s=%s by%d  opid%d\n",i,OpName[o.OpType],o.Key,o.Value,o.Who,o.OpID)
      //r+=kv.Results[i]+"\n"
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
    time.Sleep(time.Millisecond*10)
    curr:=kv.px_touchedPTR-1
    mem:=kv.snapstart
    if Debug {fmt.Printf("hosekeeper #%d, max %d, snap %d... \n",kv.me,curr,mem) }
    if(curr-mem> SaveMemThreshold){//start compressing...
      fmt.Printf("Housekeeper GC#%d starting... %d->%d\n",kv.me,mem,curr);
      kv.mu.Lock(); // Protect px.instances
        curr-=SaveMemThreshold*10/100+1
        //curr=mem+10
        if kv.snapstart==0{
          kv.snapshot=make(map[string]string)
        }
        for i:=kv.snapstart;i<curr;i++ {
          de,op:=kv.px.Status(i)
          if de==false {
            break
          }
          optt,found:=op.(Op)
          if found==false{
              println("Housekeeper error! Not Found type .(Op)")
              panic("Housekeeper sees undecided op")
          }
          k:=optt.Key
          v:=optt.Value
          bv:=kv.snapshot[k]
          switch optt.OpType{
            case GetOp:

            case PutOp:
              if bv!=""{
                //latestSucc=false
              }else{
                kv.snapshot[k]=v
                //latestSucc=true
              }

            case NaivePutOp:
              kv.snapshot[k]=v
              //latestSucc=true

            case DeleteOp:
              //latestSucc=(latestVal!="")
              //latestVal=""
              kv.snapshot[k]=""

            case UpdateOp:
              if bv==""{
                //latestSucc=false
              }else{
                kv.snapshot[k]=v
                //latestSucc=true
              }
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

    err:=kv.FormalPut(&args,&reply)
    if err!=nil || reply.Err!=""{
      fmt.Fprintf(w, "%s",kvlib.JsonErr(string(reply.Err)))
      return
    }
    fmt.Fprintf(w, "%s",kvlib.JsonSucc(reply.PreviousValue))
  }
}
//Update&Delete is similar to Put; use PaxosOps directly

func kvUpdateHandlerGC(kv *KVPaxos) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    key:= r.FormValue("key")
    value:= r.FormValue("value")
    opid:= r.FormValue("id")
    if value=="" {
      fmt.Fprintf(w, "%s",kvlib.JsonErr("value not found, please give nonempty string"))
      return
    }
    uuid:=globalOpsCnt+kv.me
    globalOpsCnt+=kv.N
    if opid!="" {
      uuid,_=strconv.Atoi(opid)
    }

    e,value:=kv.PaxosAgreementOp(Op{UpdateOp,key,value,-1,uuid})

    if e!=""{
      fmt.Fprintf(w, "%s",kvlib.JsonErr(string(e)))
      return
    }
    fmt.Fprintf(w, "%s",kvlib.JsonSucc(value))
  }
}
func kvDeleteHandlerGC(kv *KVPaxos) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    key:= r.FormValue("key")
    value:= ""
    opid:= r.FormValue("id")
    uuid:=globalOpsCnt+kv.me
    globalOpsCnt+=kv.N
    if opid!="" {
      uuid,_=strconv.Atoi(opid)
    }

    e,value:=kv.PaxosAgreementOp(Op{DeleteOp,key,value,-1,uuid})

    if e!=""{
      fmt.Fprintf(w, "%s",kvlib.JsonErr(string(e)))
      return
    }
    fmt.Fprintf(w, "%s",kvlib.JsonSucc(value))
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

    err:=kv.FormalGet(&args,&reply)
    if err!=nil || reply.Err!=""{
      fmt.Fprintf(w, "%s",kvlib.JsonErr(string(reply.Err)))
      return
    }
    fmt.Fprintf(w, "%s",kvlib.JsonSucc(reply.Value))
  }
}



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

var kvHandlerGCs = map[string]func(*KVPaxos)http.HandlerFunc{
  "insert": kvPutHandlerGC,
  "put": kvPutHandlerGC,
  "get": kvGetHandlerGC,
  "delete":kvDeleteHandlerGC,
  "update":kvUpdateHandlerGC,
}
var kvmanHandlerGCs = map[string]func(*KVPaxos)http.HandlerFunc{
  "countkey": kvmanCountKeyHandlerGC,
  "dump": kvmanDumpHandlerGC,
  "shutdown": kvmanShutdownHandlerGC,
}

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

  kv.doneOps=make(map[int]bool)
  kv.latestClientOpResult=make(map[int]ID_Ret_Pair)
  kv.Death=make(chan int,2)

  go kv.housekeeper()
  // Your initialization code here.

  if StartHTTP{

    //HTTP initialization
    //wait for a while, since previous server hasn't timed out on TCP!
    time.Sleep(time.Millisecond*11)

    serveMux := http.NewServeMux()



    for key,val := range kvHandlerGCs{
      serveMux.HandleFunc("/kv/"+key, val(kv))
    }
    for key,val := range kvmanHandlerGCs{
      serveMux.HandleFunc("/kvman/"+key, val(kv))
    }
    serveMux.HandleFunc("/", kvDumpHandlerGC(kv))

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
  fmt.Println("len is:", len(servers))
  os.Remove(servers[me])
  var socktype="unix"
  if RPC_Use_TCP==1{socktype="tcp"} // This is to help running servers between different machines!

  l, e := net.Listen(socktype, servers[me]);
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
