package kvpaxos

import "net"
import "fmt"
import "net/rpc"
import "log"
import "paxos"
import "sync"
import "os"
import "syscall"
import "encoding/gob"
import "math/rand"
import "time"
import "errors"

const Debug=0

func DPrintf(format string, a ...interface{}) (n int, err error) {
  if Debug > 0 {
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
}

func DeepCompareOps(a Op, b Op) (bool){
	return a.IsPut==b.IsPut &&
	a.Key==b.Key &&
	a.Value==b.Value &&
	a.Who==b.Who
}

type KVPaxos struct {
  mu sync.Mutex
  l net.Listener
  me int
  dead bool // for testing
  unreliable bool // for testing
  px *paxos.Paxos

  // Your definitions here.
}


func (kv *KVPaxos) Get(args *GetArgs, reply *GetReply) error {
  // Your code here.

   println("GET Step0")
  kv.mu.Lock(); // Protect px.instances
  defer kv.mu.Unlock();

   println("GET Step1")
  //step1: get the agreement!
  var myop Op
  myop.IsPut=false
  myop.Key=args.Key
  myop.Who=kv.me
  var ID int
  var value interface{}
  var decided bool
  for true {
  ID=kv.px.Max()+1
  kv.px.Start(ID,myop)
  time.Sleep(10)
  for true {
    decided,value = kv.px.Status(ID)
    if decided {
      break;
    }
    time.Sleep(100*time.Millisecond)
  }
  if DeepCompareOps(value.(Op),myop) {//succeeded
    break;
  }
  var scale=(kv.me+ID)%3
  time.Sleep(time.Duration(rand.Intn(10)*scale*int(time.Millisecond)))
  }

   println("GET Step2")
  //We got ID!
  //Step2: trace back for previous value
  var latestVal string
  var latestValFound=false

  for i:=ID-1;i>=0;i--{ //i>=latest snapshot!
  de,op:=kv.px.Status(i)
  for de==false {
    time.Sleep(10*time.Millisecond)
    de,op=kv.px.Status(i)
  }
  optt,found:=op.(Op)
  if found==false{
    return errors.New("Not Found type .(Op)")
  }
  if optt.IsPut==false{
    continue;
  }
  if optt.Key==myop.Key{
    latestVal=optt.Value
    latestValFound=true
    break
  }
  }
  if latestValFound==false {//equivalently, i<0
    reply.Err="Key Not Found"
    return nil
  }


  reply.Value=latestVal
  return nil
}

func (kv *KVPaxos) Put(args *PutArgs, reply *PutReply) error {
  // Your code here.
   println("PUT Step0")
  kv.mu.Lock(); // Protect px.instances
  defer kv.mu.Unlock();
  //step1: get the agreement!

   println("PUT Step1")
  var myop Op
  myop.IsPut=true
  myop.Key=args.Key
  myop.Value=args.Value
  myop.Who=kv.me
  var ID int
  var value interface{}
  var decided bool
  for true {
	ID=kv.px.Max()+1
  // fmt.Printf("Try to propose Min ID:%d\n",ID)

	kv.px.Start(ID,myop)
	time.Sleep(10)
	for true {
    // print("decided?")
		decided,value = kv.px.Status(ID)
		if decided {
			break;
		}
		time.Sleep(100*time.Millisecond)
	}
	if DeepCompareOps(value.(Op),myop) {//succeeded
		break;
	}
	var scale=(kv.me+ID)%3
	time.Sleep(time.Duration(rand.Intn(10)*scale*int(time.Millisecond)))
  }


   println("PUT Step2")
  //We got ID!
  //Step2: trace back for previous value
  var latestVal string
  for i:=ID-1;i>=0;i--{ //i>=latest snapshot!
	de,op:=kv.px.Status(i)
	for de==false {
	  time.Sleep(10*time.Millisecond)
	  de,op=kv.px.Status(i)
	}
  optt,found:=op.(Op)
  if found==false{
    return errors.New("Not Found type .(Op)")
  }
	if optt.IsPut==false{
	  continue;
	}
	if optt.Key==myop.Key{
	  latestVal=optt.Value
	  break
	}
  }

  // println("PUT Step3")
  // println(latestVal)
  reply.PreviousValue=latestVal
  return nil
}

// tell the server to shut itself down.
// please do not change this function.
func (kv *KVPaxos) kill() {
  DPrintf("Kill(%d): die\n", kv.me)
  kv.dead = true
  kv.l.Close()
  kv.px.Kill()
}

func (kv *KVPaxos) DumpInfo() string {
	r:=""
	r+=fmt.Sprintf("I'm %d\n",kv.me)
	r+=fmt.Sprintf("Max pxID=%d\n",kv.px.Max())
	r+=fmt.Sprintf("Min pxID=%d\n",kv.px.Min())
	ID:=kv.px.Max()
	for i=0;i<=ID;i++ {
		de,op=kv.px.Status(i)
		o,found:=op.(Op)
		if de {
			r+=fmt.Sprintf("Op[%d] IsPut%d %s=%s by%d  \n",i,o.IsPut?1:0,o.Key,o.Value,o.Who)
		}else{
			r+=fmt.Sprintf("Op[%d] undecided  \n",i)
		}
	}
}


func (kv *KVPaxos) housekeeper() {
	//launched at startup; to call paxos.done()
	/* int pointer=0, latest=0
	while(1){
		sleep for a while!
		latest=paxos latest minimum;
		for(;pointer<latest-10;pointer++)//only care about old entries
		{
			while(localcache[pointer] is missing)
			{
				try getting it from peer A,B,C
			}
			1. Myself have [pointer], know its key
			while(1){
			2. Ask peer B and C about [pointer]
			3. If A&B&C both have it, then call paxos.Done for [pointer]; break
			4. otherwise, sleep for a while, and ask again!
			}
		}
	} */
}

//HTTP handlers generator; to create a closure for kvpaxos instance
func kvDumpHandlerGC(kv *KVPaxos) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s",kv.DumpInfo())
	}
}
func kvPutHandlerGC(kv *KVPaxos) {
	return func(w http.ResponseWriter, r *http.Request) {
		key:= r.FormValue("key")
		value:= r.FormValue("value")
		fmt.Fprintf(w, "%s","")
	}
}
func kvGetHandlerGC(kv *KVPaxos) {
	return func(kv *KVPaxos, w http.ResponseWriter, r *http.Request) {
		key:= r.FormValue("key")
		fmt.Fprintf(w, "%s","")
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

  // Your initialization code here.
  
  //HTTP initialization
	var kvHandlerGCs = map[string]func(http.ResponseWriter, *http.Request){
		"dump": kvDumpHandlerGC,
		"put": kvPutHandlerGC,
		"get": kvGetHandlerGC
	}
	listenPort:=30000+me //temporary, should read from conf file!!
	s := &http.Server{
		Addr: ":"+strconv.Itoa(listenPort),
		Handler: nil,
		ReadTimeout: 1 * time.Second,
		WriteTimeout: 30 * time.Second,
		MaxHeaderBytes: 1<<20,
	}
	for key,val := range kvHandlerGCs{
		http.HandleFunc("/"+key, val(kv))
	}
	
	go func(){
		log.Fatal(s.ListenAndServe())
	}()


  
  
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

