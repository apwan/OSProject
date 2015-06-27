package kvpaxos

import (
  "net"
  "net/http"
  "net/rpc"
  "sync"
  "os"
  "syscall"
  "encoding/gob"
  "math/rand"
  "time"
  "strconv"
  "fmt"
  "log"

  "paxos"
  "stoppableHTTPlistener"
  )

const SaveMemThreshold=10000
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
  dead bool // for testing
  unreliable bool // for testing
  px *paxos.Paxos

  px_touchedPTR int

  snapshot map[string]string
  snapstart int
  
  HTTPListener *stoppableHTTPlistener.StoppableListener
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
        panic("Not decided, but before touchPTR?!!!!!")
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
func kvPutHandlerGC(kv *KVPaxos) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    key:= r.FormValue("key")
    value:= r.FormValue("value")


    var args PutArgs = PutArgs{key,value,true,-1,-1}
    var reply PutReply = PutReply{"",""}
    err:=kv.Put(&args,&reply)
    if err!=nil || reply.Err!=""{
      fmt.Fprintf(w, "{success:false,msg:%s}",reply.Err)
      return
    }
    fmt.Fprintf(w, "{success:true,value=%s}",reply.PreviousValue)
  }
}

func kvGetHandlerGC(kv *KVPaxos) http.HandlerFunc{
  return func(w http.ResponseWriter, r *http.Request) {
    key:= r.FormValue("key")

    var args GetArgs = GetArgs{key,-1,-1}
    var reply GetReply = GetReply{"",""}
    err:=kv.Get(&args,&reply)
    if err!=nil || reply.Err!=""{
      fmt.Fprintf(w, "{success:false,msg:%s}",reply.Err)
      return
    }
    fmt.Fprintf(w, "{success:true,value=%s}",reply.Value)
  }
}
//end HTTP handlers


//
// servers[] contains the ports of the set of
// servers that will cooperate via Paxos to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
//
func StartServer(servers []string, me int) *KVPaxos {
  // call gob.Register on structures you want
  // Go's RPC library to marshall/unmarshall.
  gob.Register(Op{})

  kv := new(KVPaxos)
  kv.me = me
  kv.px_touchedPTR=-1 //0 is untouched!
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
    for key,val := range kvHandlerGCs{
      serveMux.HandleFunc("/"+key, val(kv))
    }

    listenPort:=30000+me //temporary, should read from conf file!!
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
