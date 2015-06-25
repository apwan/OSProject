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
  //Get need to recover empty slots. Empty slot means 
  
  /* step1: get paxos slot
  while(1)
  {
	...
	got ID
  }
  step2: ensure no empty slot before ID
  for(i=ID-1;i>0;i--)
  {
	while(localcache[i]) empty
	{
		try from peer A, skip if I'm A
		try from peer B, skip if...
		try from peer C, ...
	}
	if(localcache[i].key==key)
	{
		latestvalue=localcache[i].value;
		break;
	}
  }
  step3: fill information
  localcache[ID]=
	type=get, key=key, value=val, 
  return nil */
  
  
}

func (kv *KVPaxos) Put(args *PutArgs, reply *PutReply) error {
  // Your code here.
  kv.mu.Lock(); // Protect px.instances
  defer kv.mu.Unlock();
  //step1: get the agreement!
  var myop Op
  myop.IsPut=true
  myop.Key=args.Key
  myop.Value=args.Value
  myop.Who=kv.me
  var ID int
  var value Op
  var decided bool
  for true {
	ID=kv.px.Min()
	kv.px.Start(ID,myop)
	time.Sleep(10)
	for true {
		decided,value = kv.px.Status(ID)
		if decided {
			break;
		}
		time.Sleep(100*time.Millisecond())
	}
	if value==myop {//succeeded
		break;
	}
	var scale=(kv.me+ID)%3
	time.Sleep((rand.Int63() % (scale*100))*time.Millisecond())
  }
  //We got ID!
  //Step2: trace back for previous value
  var latestVal string
  for i:=ID-1;i>=0;i--{ //i>=latest snapshot!
	de,op:=kv.px.Status(i)
	while(de==false){
	  time.Sleep(10*time.Millisecond())
	  de,op=kv.px.Status(i)
	}
	if op.IsPut==false{
	  continue;
	}
	if op.Key==myop.Key{
	  latestVal=op.Value
	  break
	}
  }
  PutReply.PreviousValue=latestVal
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

