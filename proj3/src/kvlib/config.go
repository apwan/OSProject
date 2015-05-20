package kvlib

import(
  "os"
  "fmt"
  "strconv"
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
