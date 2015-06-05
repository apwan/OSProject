package main

import (
    "fmt"
    "os"
    "os/exec"
    "net/http"
    "net/url"
    "time"
    "strconv"
    // our lib
    . "kvlib"
)




func checkDump(t map[string]string, d map[string]interface{}) int {
    if len(t) != len(d) {
        return 1
    }
    for k, v := range t {
        if d[k] != v {
            return 1
        }
    }
    return 0
}




func TestUnit(p, b, fn string) (r string, fail int) {
    f, _ := os.Open(fn)
    table := make(map[string]string)
    var tableBlock map[string]int
    var s [4]string
    var inBlock, primary, backup, cnt, cins int
    var ch chan int
    var ins chan [4]string
    r += fmt.Sprintf("Testing %s..\n\n", fn)
    for {
        l, _ := fmt.Fscanln(f, &s[0], &s[1], &s[2], &s[3])
        if l == 0 {
            if primary == 1 {
                resp, err := http.Get(p + "/kvman/dump")
                if err != nil {
                    r += fmt.Sprintf("Dumping failed.\n")
                    r += fmt.Sprintf("FATAL ERROR!!!\n")
                    fail = 1
                } else {
                    dump := DecodeJson(resp)
                    if checkDump(table, dump) != 0 {
                        r += fmt.Sprintf("Incorrect primary dump.\n")
                        r += fmt.Sprintf("Expected: %v, dumped: %v\n", table, dump)
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    }
                }
            }
            if backup == 1 {
                resp, err := http.Get(b + "/kvman/dump")
                if err != nil {
                    r += fmt.Sprintf("Dumping failed.\n")
                    r += fmt.Sprintf("FATAL ERROR!!!\n")
                    fail = 1
                } else {
                    dump := DecodeJson(resp)
                    if checkDump(table, dump) != 0 {
                        r += fmt.Sprintf("Incorrect backup dump.\n")
                        r += fmt.Sprintf("Expected: %v, dumped: %v\n", table, dump)
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    }
                }
            }
            http.Get(p + "/kvman/shutdown")
            http.Get(b + "/kvman/shutdown")
            r += fmt.Sprintf("\nTerminating.\n")
            return
        }
        for i := 0; i < l; i++ {
            if i != 0 {
                r += fmt.Sprintf(" %v", s[i])
            } else {
                r += fmt.Sprintf("\nInstruction: %v", s[i])
            }
        }
        for i := l; i < 4; i++ {
            s[i] = ""
        }
        r += fmt.Sprintf("\nResult: ")
        switch s[0] {
            case "Insert":
                if inBlock == 1 {
                    if tableBlock[s[1]] == 1 {
                        r += fmt.Sprintf("Illegal instruction: key appeared twice in a block.\n")
                    } else {
                        tableBlock[s[1]] = 1
                        cnt++
                        if _, ok := table[s[1]]; ok == false && primary == 1 && backup == 1 {
                            cins++
                            ins <- s
                        }
                        go func(s [4]string) {
                            resp, err := http.PostForm(p + "/kv/insert", url.Values{"key": {s[1]}, "value": {s[2]}})
                            if err == nil {
                                res := DecodeJson(resp)
                                if res["success"] == "true" {
                                    if backup == 0 {
                                        ch <- 1
                                    } else if _, ok := table[s[1]]; ok == true {
                                        ch <- 1
                                    } else {
                                        ch <- 0
                                    }
                                } else if backup == 0 {
                                    ch <- 0
                                } else if _, ok := table[s[1]]; ok == false {
                                    ch <- 1
                                } else {
                                    ch <- 0
                                }
                            } else {
                                if primary == 0 {
                                    ch <- 0
                                } else {
                                    ch <- 1
                                }
                            }
                        }(s)
                    }
                    break
                }
                resp, err := http.PostForm(p + "/kv/insert", url.Values{"key": {s[1]}, "value": {s[2]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {
                        if backup == 0 {
                            r += fmt.Sprintf("Backup is down.\n")
                            r += fmt.Sprintf("Unexpected insertion success.\n")
                            r += fmt.Sprintf("FATAL ERROR!!!\n")
                            fail = 1
                        } else if _, ok := table[s[1]]; ok == true {
                            r += fmt.Sprintf("Unexpected insertion success.\n")
                            r += fmt.Sprintf("FATAL ERROR!!!\n")
                            fail = 1
                        } else {
                            r += fmt.Sprintf("Insertion success.\n")
                        }
                    } else if backup == 0 {
                        r += fmt.Sprintf("Backup is down.\n")
                        r += fmt.Sprintf("Expected insertion failure.\n")
                    } else if _, ok := table[s[1]]; ok == false {
                        r += fmt.Sprintf("Unexpected insertion failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    } else {
                        r += fmt.Sprintf("Expected insertion failure.\n")
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                    if primary == 0 {
                        r += fmt.Sprintf("Primary is down.\n")
                    } else {
                        r += fmt.Sprintf("Primary does not response.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    }
                }
                if _, ok := table[s[1]]; ok == false && primary == 1 && backup == 1 {
                    table[s[1]] = s[2]
                }
            case "Update":
                if inBlock == 1 {
                    if tableBlock[s[1]] == 1 {
                        r += fmt.Sprintf("Illegal instruction: key appeared twice in a block.\n")
                    } else {
                        tableBlock[s[1]] = 1
                        cnt++
                        if _, ok := table[s[1]]; ok == true && primary == 1 && backup == 1 {
                            cins++
                            ins <- s
                        }
                        go func(s [4]string) {
                            resp, err := http.PostForm(p + "/kv/update", url.Values{"key": {s[1]}, "value": {s[2]}})
                            if err == nil {
                                res := DecodeJson(resp)
                                if res["success"] == "true" {
                                    if backup == 0 {
                                        ch <- 1
                                    } else if _, ok := table[s[1]]; ok == false {
                                        ch <- 1
                                    } else {
                                        ch <- 0
                                    }
                                } else if backup == 0 {
                                    ch <- 0
                                } else if _, ok := table[s[1]]; ok == true {
                                    ch <- 1
                                } else {
                                    ch <- 0
                                }
                            } else {
                                if primary == 0 {
                                    ch <- 0
                                } else {
                                    ch <- 1
                                }
                            }
                        }(s)
                    }
                    break
                }
                resp, err := http.PostForm(p + "/kv/update", url.Values{"key": {s[1]}, "value": {s[2]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {
                        if backup == 0 {
                            r += fmt.Sprintf("Backup is down.\n")
                            r += fmt.Sprintf("Unexpected updating success.\n")
                            r += fmt.Sprintf("FATAL ERROR!!!\n")
                            fail = 1
                        } else if _, ok := table[s[1]]; ok == false {
                            r += fmt.Sprintf("Unexpected updating success.\n")
                            r += fmt.Sprintf("FATAL ERROR!!!\n")
                            fail = 1
                        } else {
                            r += fmt.Sprintf("Updating success.\n")
                        }
                    } else if backup == 0 {
                        r += fmt.Sprintf("Backup is down.\n")
                        r += fmt.Sprintf("Expected updating failure.\n")
                    } else if _, ok := table[s[1]]; ok == true {
                        r += fmt.Sprintf("Unexpected updating failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    } else {
                        r += fmt.Sprintf("Expected updating failure.\n")
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                    if primary == 0 {
                        r += fmt.Sprintf("Primary is down.\n")
                    } else {
                        r += fmt.Sprintf("Primary does not response.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    }
                }
                if _, ok := table[s[1]]; ok == true && primary == 1 && backup == 1 {
                    table[s[1]] = s[2]
                }
            case "Remove":
                if inBlock == 1 {
                    if tableBlock[s[1]] == 1 {
                        r += fmt.Sprintf("Illegal instruction: key appeared twice in a block.\n")
                    } else {
                        tableBlock[s[1]] = 1
                        cnt++
                        if _, ok := table[s[1]]; ok == true && primary == 1 && backup == 1 {
                            cins++
                            ins <- s
                        }
                        go func(s [4]string) {
                            resp, err := http.PostForm(p + "/kv/delete", url.Values{"key": {s[1]}})
                            if err == nil {
                                res := DecodeJson(resp)
                                if res["success"] == "true" {
                                    if backup == 0 {
                                        ch <- 1
                                    } else if _, ok := table[s[1]]; ok == false {
                                        ch <- 1
                                    } else {
                                        if res["value"] == table[s[1]] {
                                            ch <- 0
                                        } else {
                                            ch <- 1
                                        }
                                    }
                                } else if backup == 0 {
                                    ch <- 0
                                } else if _, ok := table[s[1]]; ok == true {
                                    ch <- 1
                                } else {
                                    ch <- 0
                                }
                            } else {
                                if primary == 0 {
                                    ch <- 0
                                } else {
                                    ch <- 1
                                }
                            }
                        }(s)
                    }
                    break
                }
                resp, err := http.PostForm(p + "/kv/delete", url.Values{"key": {s[1]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {
                        if backup == 0 {
                            r += fmt.Sprintf("Backup is down.\n")
                            r += fmt.Sprintf("Unexpected deleting success.\n")
                            r += fmt.Sprintf("FATAL ERROR!!!\n")
                            fail = 1
                        } else if _, ok := table[s[1]]; ok == false {
                            r += fmt.Sprintf("Unexpected deleting success.\n")
                            r += fmt.Sprintf("FATAL ERROR!!!\n")
                            fail = 1
                        } else {
                            r += fmt.Sprintf("Deleting success.\n")
                            if res["value"] == table[s[1]] {
                                r += fmt.Sprintf("Correct value deleted.\n")
                            } else {
                                r += fmt.Sprintf("Incorrect value deleted.\n")
                                r += fmt.Sprintf("FATAL ERROR!!!\n")
                                fail = 1
                            }
                        }
                    } else if backup == 0 {
                        r += fmt.Sprintf("Backup is down.\n")
                        r += fmt.Sprintf("Expected deleting failure.\n")
                    } else if _, ok := table[s[1]]; ok == true {
                        r += fmt.Sprintf("Unexpected deleting failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    } else {
                        r += fmt.Sprintf("Expected deleting failure.\n")
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                    if primary == 0 {
                        r += fmt.Sprintf("Primary is down.\n")
                    } else {
                        r += fmt.Sprintf("Primary does not response.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    }
                }
                if _, ok := table[s[1]]; ok == true && primary == 1 && backup == 1 {
                    delete(table, s[1])
                }
            case "Get":
                if inBlock == 1 {
                    if tableBlock[s[1]] == 1 {
                        r += fmt.Sprintf("Illegal instruction: key appeared twice in a block.\n")
                    } else {
                        tableBlock[s[1]] = 1
                        go func(s [4]string) {
                            resp, err := http.Get(p + "/kv/get?key=" + s[1])
                            if err == nil {
                                res := DecodeJson(resp)
                                if res["success"] == "true" {
                                    if _, ok := table[s[1]]; ok == false {
                                        ch <- 1
                                    } else {
                                        if res["value"] == table[s[1]] {
                                            ch <- 0
                                        } else {
                                            ch <- 1
                                        }
                                    }
                                } else if backup == 0 {
                                    ch <- 0
                                } else if _, ok := table[s[1]]; ok == true {
                                    ch <- 1
                                } else {
                                    ch <- 0
                                }
                            } else {
                                if primary == 0 {
                                    ch <- 0
                                } else {
                                    ch <- 1
                                }
                            }
                        }(s)
                    }
                    break
                }
                resp, err := http.Get(p + "/kv/get?key=" + s[1])
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {
                        if _, ok := table[s[1]]; ok == false {
                            r += fmt.Sprintf("Unexpected getting success.\n")
                            fail = 1
                        } else {
                            r += fmt.Sprintf("Getting success.\n")
                            if res["value"] == table[s[1]] {
                                r += fmt.Sprintf("Got correct value.\n")
                            } else {
                                r += fmt.Sprintf("Got incorrect value.\n")
                                r += fmt.Sprintf("FATAL ERROR!!!\n")
                                fail = 1
                            }
                        }
                    } else if backup == 0 {
                        r += fmt.Sprintf("Backup is down.\n")
                        r += fmt.Sprintf("Expected getting failure.\n")
                    } else if _, ok := table[s[1]]; ok == true {
                        r += fmt.Sprintf("Unexpected getting failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    } else {
                        r += fmt.Sprintf("Expected getting failure.\n")
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                    if primary == 0 {
                        r += fmt.Sprintf("Primary is down.\n")
                    } else {
                        r += fmt.Sprintf("Primary does not response.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    }
                }
            case "Exec":
                if inBlock == 0 {
                    flag := 1
                    go CmdWaiter(exec.Command(s[1], s[2], s[3]), &r, &flag)
                    time.Sleep(1000*time.Millisecond)
                    if flag == 1{
                      r += fmt.Sprintf("Exec time exceed!\n")
                      flag = 0
                      //fail = 1 // allowed
                    }

                } else {
                    r += fmt.Sprintf("Illegal instruction: Exec in a block.\n")
                }
            case "Sleep":
                if inBlock == 0 {
                    t, _ := strconv.Atoi(s[1])
                    time.Sleep(time.Duration(t) * time.Millisecond)
                    r += fmt.Sprintf("Slept for %v milliseconds.\n", t)
                } else {
                    r += fmt.Sprintf("Illegal instruction: Sleep in a block.\n")
                }
            case "Block":
                inBlock = 1
                tableBlock = make(map[string]int)
                ch = make(chan int, 10000)
                ins = make(chan [4]string, 10000)
                cnt = 0
                cins = 0
                r += fmt.Sprintf("Entering a block.\n")
            case "Endblock":
                r += fmt.Sprintf("Leaving a block of %v legal instructions.\n", cnt)
                for i := 0; i < cnt; i++ {
                    if v := <-ch; v != 0 {
                        r += fmt.Sprintf("Something went wrong in the block.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    }
                }
                for i := 0; i < cins; i++ {
                    s = <-ins
                    switch s[0] {
                        case "Insert", "Update":
                            table[s[1]] = s[2]
                        case "Remove":
                            delete(table, s[1])
                    }
                }
                inBlock = 0
            case "Switch":
                if s[1] == "primary" {
                    primary ^= 1
                    r += fmt.Sprintf("Primary state: %v\n", primary)
                } else {
                    backup ^= 1
                    r += fmt.Sprintf("Backup state: %v\n", backup)
                }
                if primary == 0 && backup == 0 {
                    table = make(map[string]string)
                }
            default:
                r += fmt.Sprintf("Unrecognised instruction.\n")
        }
    }
}



func main() {
    confname := "conf/test.conf";
    if len(os.Args)>1 && os.Args[1]!="-direct" {
      confname = os.Args[1];
    }
    conf := ReadJson(confname);
    primary := "http://" + conf["primary"]
    backup := "http://" + conf["backup"]
    tot,_ := strconv.Atoi(conf["total"])
    cnt := tot

    jump := false
    if len(os.Args)>1 && os.Args[1]=="-direct"{
      // without starting/stoping server, only test the performance
      jump = true
      tot = 0
      cnt = 0
    }

    for i := 0; i < tot; i++ {
      testname := conf["pre"]+strconv.Itoa(i)+".test"

      if conf["fmt"] != "true"{ //no need to specify each test case name
        testname = conf["pre"]+conf[strconv.Itoa(i)]
      }
        res, fail := TestUnit(primary, backup, testname)
        if conf["with_err_msg"]=="true"{
          fmt.Printf("%s", res)
          if fail == 0 {
              fmt.Printf("\nTest case %d: success!\n\n", i)
          } else {
              fmt.Printf("\nTest case %d: failed!\n\n", i)
          }


        }
        cnt -= fail

    }


    if cnt == tot {
        fmt.Printf("success\n")
    } else {
        fmt.Printf("fail\n")
    }

    kvURL := primary+"/kv/"


    if !jump{
      StartServer("-p")
      StartServer("-b")
    }

    time.Sleep(500*time.Millisecond)
    N,err := strconv.Atoi(conf["concur_num"])
    if err != nil {
      N = 125
    }

    TestPerformance(N, kvURL)

    if !jump{
      StopServer("-p")
      StopServer("-b")
    }



}
