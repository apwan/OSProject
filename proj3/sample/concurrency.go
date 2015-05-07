package main

import (
  "fmt"
  "time"
)

func pinger(c chan string) {
    for i := 0; ; i++ {
        c <- "ping"
    }
}
func ponger(c chan string) {
    for i := 0; ; i++ {
        c <- "pong"
    }
}
func printer(c chan string) {
    for {
        msg := <- c
        fmt.Println(msg)
        time.Sleep(time.Second * 1)
    }
}



func main(){
  fmt.Println("enter any key to exit ...")
  var c1 chan string = make(chan string)
  c2 := make(chan string)

    go pinger(c1)
    go ponger(c2)
    //go printer(c1)
    go func() {
        for {
            select {
            case msg1 := <- c1:
                fmt.Println(msg1)
            case msg2 := <- c2:
                fmt.Println(msg2)
            }
            time.Sleep(time.Second * 1)
        }
    }()

    var input string
    fmt.Scanln(&input)
}
