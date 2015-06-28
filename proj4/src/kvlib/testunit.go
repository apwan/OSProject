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

func TestUnit(addr [3]string, fn string) (r string, fail int) {
    f, _ := os.Open(fn)
    table := make(map[string]string)
    var tableBlock map[string]int
    var s [4]string
    var srv_cur int
    var inBlock, cnt, cins int
    var ch chan int
    var ins chan [4]string
    r += fmt.Sprintf("Testing %s..\n\n", fn)
    for {
        l, _ := fmt.Fscanln(f, &s[0], &s[1], &s[2], &s[3])

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
                        if _, ok := table[s[1]]; ok == false{
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
                    }
                    break
                }
                resp, err := http.PostForm(addr[srv_cur] + "/kv/insert", url.Values{"key": {s[1]}, "value": {s[2]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {

                        r += fmt.Sprintf("Insertion success.\n")

                    } else  if _, ok := table[s[1]]; ok == false {
                        r += fmt.Sprintf("Unexpected insertion failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    } else {
                        r += fmt.Sprintf("Expected insertion failure.\n")
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                }
                if _, ok := table[s[1]]; ok == false{
                    table[s[1]] = s[2]
                }
            case "Update":
                if inBlock == 1 {
                    if tableBlock[s[1]] == 1 {
                        r += fmt.Sprintf("Illegal instruction: key appeared twice in a block.\n")
                    } else {
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
                    }
                    break
                }
                resp, err := http.PostForm(addr[srv_cur] + "/kv/update", url.Values{"key": {s[1]}, "value": {s[2]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {

                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")

                }
                if _, ok := table[s[1]]; ok == true {
                    table[s[1]] = s[2]
                }
            case "Remove":
                if inBlock == 1 {
                    if tableBlock[s[1]] == 1 {
                        r += fmt.Sprintf("Illegal instruction: key appeared twice in a block.\n")
                    } else {
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
                    }
                    break
                }
                resp, err := http.PostForm(addr[srv_cur] + "/kv/delete", url.Values{"key": {s[1]}})
                if err == nil {
                    res := DecodeJson(resp)
                    r += fmt.Sprintf("%v\n", res)
                    if res["success"] == "true" {

                            r += fmt.Sprintf("Deleting success.\n")
                            if res["value"] == table[s[1]] {
                                r += fmt.Sprintf("Correct value deleted.\n")
                            } else {
                                r += fmt.Sprintf("Incorrect value deleted.\n")
                                r += fmt.Sprintf("FATAL ERROR!!!\n")
                                fail = 1
                            }

                    }else if _, ok := table[s[1]]; ok == true {
                        r += fmt.Sprintf("Unexpected deleting failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    } else {
                        r += fmt.Sprintf("Expected deleting failure.\n")
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                }
                if _, ok := table[s[1]]; ok == true{
                    delete(table, s[1])
                }
            case "Get":
                if inBlock == 1 {
                    if tableBlock[s[1]] == 1 {
                        r += fmt.Sprintf("Illegal instruction: key appeared twice in a block.\n")
                    } else {
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
                    }
                    break
                }
                resp, err := http.Get(addr[srv_cur] + "/kv/get?key=" + s[1])
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
                    } else if _, ok := table[s[1]]; ok == true {
                        r += fmt.Sprintf("Unexpected getting failure.\n")
                        r += fmt.Sprintf("FATAL ERROR!!!\n")
                        fail = 1
                    } else {
                        r += fmt.Sprintf("Expected getting failure.\n")
                    }
                } else {
                    r += fmt.Sprintf("Error occurred.\n")
                }
            case "Exec":
                if inBlock == 0 {

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
                tmp,_ := strconv.Atoi(s[1])
                srv_cur = tmp-1
                r += fmt.Sprintf("Switch to Server: %d\n", srv_cur)

            default:
                r += fmt.Sprintf("Unrecognised instruction.\n")
        }
    }
}
