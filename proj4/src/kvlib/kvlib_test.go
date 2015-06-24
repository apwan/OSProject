package kvlib

import(
  "fmt"
  "time"
  "testing"
)
func Test1(t *testing.T) {
  fin := "in.json"
  fout := "out.json"
  dat := ReadJson(fin)
  err := WriteJson(fout, dat)
  if err != nil {
    fmt.Println(err)
    return
  }
}
