package kvlib

import (
    "fmt"
    "os"
    //"os/exec"
    "net/http"
    "net/url"
    "time"
    "strconv"

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

func TestUnit(addr [3]string, tester_addr [3]string, fn string) (r string, fail int) {
    f, _ := os.Open(fn)
    defer f.Close()
    table := make(map[string]string)
    var tableBlock map[string]int
    var s [4]string
    var srv_cur, livingServer int
    var alive [3]int
    var inBlock, cnt, cins int
    inBlock = 0
    var ch chan int
    var ins chan [4]string
    r += fmt.Sprintf("Testing %s..\n\n", fn)
    fmt.Println(r)
    for {
        l, _ := fmt.Fscanln(f, &s[0], &s[1], &s[2], &s[3])
        if l==0{
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
        case "Put":
          fmt.Println("PUT")
                if inBlock == 1 {
                    tableBlock[s[1]] = 1
                    cnt++
                    if _, ok := table[s[1]]; ok == false {
                        cins++
                        ins <- s
                    }
                    go func(s [4]string) {
                        resp, err := http.PostForm(addr[srv_cur] + "/kv/insert", url.Values{"key": {s[1]}, "value": {s[2]}})
                        if err == nil {
                            res := DecodeJson(resp)
                            if res["success"] == "true" {

                            }
                        } else {

                        }
                    }(s)
                    break
                }
                resp, err := http.PostForm(addr[srv_cur] + "/kv/insert", url.Values{"key": {s[1]}, "value": {s[2]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {
                        if _, ok := table[s[1]]; ok == false {
                            r += fmt.Sprintf("Insertion success.\n")
                            table[s[1]] = s[2]
                        } else {
                            r += fmt.Sprintf("Unexpected insertion success.\n")
                            r += fmt.Sprintf("FATAL ERROR!!!\n")
                            fail = 1
                        }
                    } else if _, ok := table[s[1]]; ok == false {
                        r += fmt.Sprintf("Unexpected insertion failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    } else {
                        r += fmt.Sprintf("Expected insertion failure.\n")
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                }
            case "Update":
              fmt.Println("Update")
                if inBlock == 1 {
                    tableBlock[s[1]] = 1
                    cnt++
                    if _, ok := table[s[1]]; ok == true {
                        cins++
                        ins <- s
                    }
                    go func(s [4]string) {
                        resp, err := http.PostForm(addr[srv_cur] + "/kv/update", url.Values{"key": {s[1]}, "value": {s[2]}})
                        if err == nil {
                            res := DecodeJson(resp)
                            if res["success"] == "true" {


                            }
                        } else {

                        }
                    }(s)
                    break
                }
                resp, err := http.PostForm(addr[srv_cur] + "/kv/update", url.Values{"key": {s[1]}, "value": {s[2]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {
                        if _, ok := table[s[1]]; ok == true {
                            r += fmt.Sprintf("Updating Success.\n")
                            table[s[1]] = s[2]
                        } else {
                            r += fmt.Sprintf("Unexpected updating success.\n")
                            r += fmt.Sprintf("FATAL ERROR!!!\n")
                            fail = 1
                        }
                    } else if _, ok := table[s[1]]; ok == true {
                            r += fmt.Sprintf("Expected updating failure.\n")
                    } else {
                        r += fmt.Sprintf("Unexpected updating failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                }
            case "Delete":
              fmt.Println("DELETE")
                if inBlock == 1 {
                    tableBlock[s[1]] = 1
                    cnt++
                    if _, ok := table[s[1]]; ok == true{
                        cins++
                        ins <- s
                    }
                    go func(s [4]string) {
                        resp, err := http.PostForm(addr[srv_cur] + "/kv/delete", url.Values{"key": {s[1]}})
                        if err == nil {
                            res := DecodeJson(resp)
                            if res["success"] == "true" {

                            }
                        } else {

                        }
                    }(s)
                    break
                }
                resp, err := http.PostForm(addr[srv_cur] + "/kv/delete", url.Values{"key": {s[1]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {
                        if _, ok := table[s[1]]; ok == true {
                            r += fmt.Sprintf("Deleting success.\n")
                            delete(table, s[1])
                            if res["value"] == table[s[1]] {
                                r += fmt.Sprintf("Correct value deleted.\n")
                            } else {
                                r += "Incorrect value deleted.\n"+
                                     "FATAL ERROR!!!\n"
                                fail = 1
                            }
                        } else {
                            r += "Unexpected deleting success.\n"+
                                 "FATAL ERROR!!!\n"
                            fail = 1
                        }
                    } else if _, ok := table[s[1]]; ok == true {
                        r += "Unexpected deleting failure.\n"+
                             "FATAL ERROR!!!\n"
                        fail = 1
                    } else {
                        r += "Expected deleting failure.\n"
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                }
            case "Get":
              fmt.Println("GET")
                if inBlock == 1 {
                    tableBlock[s[1]] = 1
                    go func(s [4]string) {
                        resp, err := http.Get(addr[srv_cur] + "/kv/get?key=" + s[1])
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
                            } else if _, ok := table[s[1]]; ok == true {
                                ch <- 1
                            } else {
                                ch <- 0
                            }
                        } else {

                        }
                    }(s)
                    break
                }else{


                resp, err := http.Get(addr[srv_cur] + "/kv/get?key=" + s[1])
                if err != nil {
                    r += "Error occurred.\n"
                }else{
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    fmt.Println("Decode Json")
                    if res["success"] == "true" {
                        if _, ok := table[s[1]]; ok == false {
                            r += "Unexpected getting success.\n"
                            fail = 1
                        } else {
                            r += "Getting success.\n"
                            if res["value"] == table[s[1]] {
                                r += "Got correct value.\n"
                            } else {
                                r += "Got incorrect value.\n"+
                                     "FATAL ERROR!!!\n"
                                fail = 1
                            }
                        }
                    } else if _, ok := table[s[1]]; ok == true {
                        r += "Unexpected getting failure.\n"+
                             "FATAL ERROR!!!\n"
                        fail = 1
                    } else {
                        r += "Expected getting failure.\n"
                    }
                }
              }
            /*
            case "Exec":
                if inBlock == 0 {

                } else {
                    r += fmt.Sprintf("Illegal instruction: Exec in a block.\n")
                }
            */
            case "Sleep":
              fmt.Println("Sleep")
                if inBlock == 0 {
                    t, _ := strconv.Atoi(s[1])
                    time.Sleep(time.Duration(t) * time.Millisecond)
                    r += fmt.Sprintf("Slept for %v milliseconds.\n", t)
                } else {
                    r += "Illegal instruction: Sleep in a block.\n"
                }
            case "Block":
                inBlock = 1
                tableBlock = make(map[string]int)
                ch = make(chan int, 10000)
                ins = make(chan [4]string, 10000)
                cnt = 0
                cins = 0
                r += "Entering a block.\n"
            case "Endblock":
                r += fmt.Sprintf("Leaving a block of %v legal instructions.\n", cnt)
                for i := 0; i < cnt; i++ {
                    if v := <-ch; v != 0 {
                        r += "Something went wrong in the block.\n"+
                             "FATAL ERROR!!!\n"
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
              fmt.Println("Switch")
                tmp, _ := strconv.Atoi(s[1])
                srv_cur = tmp - 1
                r += fmt.Sprintf("Switch to Server: %d\n", srv_cur)
            case "start_server":
                t, _ := strconv.Atoi(s[1])
                if alive[t - 1] == 0 {
                    livingServer++
                    alive[t - 1] = 1
                    resp, err := http.Get(tester_addr[t-1] + "/test/start_server")
                    if err != nil{
                      fmt.Println(err)
                    }else{
                      fmt.Println(DecodeStr(resp))
                    }
                    //exec.Run(exec.Command("bin/start_server", fmt.Sprintf("%v", t)))
                }
            case "stop_server":
                t, _ := strconv.Atoi(s[1])
                if alive[t - 1] == 1 {
                    livingServer--
                    alive[t - 1] = 0
                    resp, err := http.Get(tester_addr[t-1] + "/test/stop_server")
                    if err != nil{
                      fmt.Println(err)
                    }else{
                      fmt.Println(DecodeStr(resp))
                    }
                    //exec.Run(exec.Command("bin/stop_server", fmt.Sprintf("%v", t)))
                }
            default:
                r += "Unrecognised instruction.\n"
        }
    }
}
