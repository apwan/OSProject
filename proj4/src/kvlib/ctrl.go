package kvlib

import(
  "fmt"
  "os/exec"
  "time"
)

func CmdWaiter(cmd *exec.Cmd, rr *string, flag *int){
  cmd.Start()
  err := cmd.Wait()
  if *flag == 0 {
    return
  }
  if err == nil {
      *rr += fmt.Sprintf("Exec Success.\n")
  } else {
      *rr += fmt.Sprintf("Exec Error: %v\n", err)
  }
  *flag = 0
}

func CmdStartServer(arg string)string{
  flag := 1
  r := ""
  go CmdWaiter(exec.Command("bin/start_server",arg), &r, &flag)
  time.Sleep(500*time.Millisecond)
  if flag == 1{
    r += "Time Exceed"
    flag = 0
  }
  return r

}

func CmdStopServer(arg string)string{
  cmd := exec.Command("bin/stop_server",arg);
  err := cmd.Run()
  if err != nil{
    return "Failed"
  }else{
    return "Success"
  }
}
