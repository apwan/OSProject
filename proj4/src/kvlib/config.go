package kvlib

import(
  //"os"
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
	p,err := strconv.Atoi(conf["port_n"+strconv.Itoa(me)])
		if err != nil {
			fmt.Println("Failed to parse :"+"port_n"+strconv.Itoa(me)+":"+conf["port_n"+strconv.Itoa(me)]);
			panic(err)
		}
	return p
}
