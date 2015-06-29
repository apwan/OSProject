package kvlib

import(
  "os"
  "fmt"
  "strconv"
  //"time"
  //"math/rand"
  "encoding/json"
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
type MsgResponse struct {
	Success string `json:"success"`
    Message string `json:"message"`
}

func JsonErr(Err string)(string){
  enc,_:=json.Marshal(&MsgResponse{"false",Err});
  return string(enc)
}
func JsonSucc(Val string)(string){
  enc,_:=json.Marshal(&StrResponse{"true",Val});
  return string(enc)
}

func Find_Port(me int, conf map[string]string) (int){
  id := fmt.Sprintf("n%02d", me+1)
	if conf["use_different_port"] == "true" {
    p,err := strconv.Atoi(conf["port_"+id])
  		if err != nil {
  			fmt.Printf("Failed to parse :"+"port_%s : %s\n", id, conf["port_"+id]);
  			panic(err)
  		}
  	return p	
	}
  p,err := strconv.Atoi(conf["port"])
  if err != nil {
    println("Failed to parse conf[port]")
    panic(err)
  }
  return p


}

func Det_role() int {
	arg_num := len(os.Args)
  if arg_num <= 1 {
    return -1 //invalid role
  }
  if os.Args[1]=="-m"{ // for main tester
    return 0
  }
  tmp := []rune(os.Args[1])
  if len(tmp)!=3 || tmp[0]!='n'{
    a,e := strconv.Atoi(os.Args[1])
    if e==nil{
      return a
    }else{
      return -1
    }
  }
  ret,err := strconv.Atoi(string(tmp[1:]))
  if(err!=nil){
    return -1
  }else{
    return ret
  }
}
