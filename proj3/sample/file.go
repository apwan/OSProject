package main

import(
  "bufio"
  "os"
  "fmt"
  "io"
  )

func main(){
  inputFile, ierr := os.Open(os.Args[1])
  if ierr != nil {
    fmt.Println("input file err!")
    return
  }
  defer inputFile.Close()

  inputReader := bufio.NewReader(inputFile)
  lineCount := 0
  for{
    inputString, rerr := inputReader.ReadString('\n')
    if rerr == io.EOF {
      return
    }
    lineCount++
    fmt.Printf("%d : %s", lineCount, inputString)
  }

}
