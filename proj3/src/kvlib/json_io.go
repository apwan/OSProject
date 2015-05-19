package kvlib

import(
  "encoding/json"
  "io/ioutil"
)

func ReadJson(s string) (map[string]string){
  dat, err := ioutil.ReadFile(s)
	if err != nil {
        panic(err)
  }

	var udat map[string]interface{}
	if err := json.Unmarshal(dat, &udat); err != nil {
        panic(err)
    }
	ret:=make(map[string]string)

	for key, val := range udat {
		var str=val.(string)
		ret[key]=str
    }

	return ret
}

func WriteJson(s string, udat map[string]string)(error){

  dat, err := json.Marshal(udat)
  if err != nil {
    panic(err)
  }

  if err := ioutil.WriteFile(s, dat, 0644); err != nil {
    panic(err)
  }
  return err
}
