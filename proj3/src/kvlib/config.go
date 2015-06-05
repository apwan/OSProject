package kvlib

import(
  "os"
  "fmt"
  "strconv"
  "time"
  "math/rand"
)

const(
 PRIMARY=1
 BACKUP=2
)
const(
 COLD_START=0
 WARM_START=1
 BOOTSTRAP=2
 SYNC=3
 SHUTTING_DOWN=-1
)

type BoolResponse struct {
   Success bool `json:"success"`
}
var (
 TrueResponseStr = "{\"success\":\"true\"}"
 FalseResponseStr = "{\"success\":\"false\"}"
)// in high-performance setting, TRS="1", FRS="0" !!!

type StrResponse struct {
	Success string `json:"success"`
    Value string `json:"value"`
}

type Duration_slice []time.Duration
func (a Duration_slice) Len() int { return len(a) }
func (a Duration_slice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Duration_slice) Less(i, j int) bool { return a[i] < a[j] }

func GenLongStr()(string){
  dummy:="TEST kv long str ......"
  for i:=0; i<10;i++ {
	dummy=dummy+ string(i%26+65)
  }
  r := rand.New(rand.NewSource(time.Now().UnixNano()))
  dummy=fmt.Sprintf("rand%lf", r.Float64())+dummy+fmt.Sprintf("rand%lf", r.Float64())+":"
  return dummy
}

func Det_role() int {
	arg_num := len(os.Args)
	for i := 0 ; i < arg_num ;i++{
		switch os.Args[i] {
			case "-p":
				return PRIMARY
			case "-b":
				return BACKUP
		}
	}
  fmt.Println("Unknown; please specify role as command line parameter.")
  panic(os.Args)
	return 0
}
func Find_port(role int, conf map[string]string) (int,int,int){
	p,err := strconv.Atoi(conf["port"])
		if err != nil {
			fmt.Println("Failed to parse port:"+conf["port"]);
			panic(err)
		}
	bp,err := strconv.Atoi(conf["back_port"])
		if err != nil {
			fmt.Println("Failed to parse back_port:"+conf["back_port"]);
			panic(err)
		}
		if (bp == 0 ){
      fmt.Println("Invalid back_port:");
			panic(conf["back_port"])
		}

	if conf["primary"] != conf["backup"]{
		return p,p,p
	}

	if role==PRIMARY{
		return p,p,bp
	}
	return bp,p,bp
}
