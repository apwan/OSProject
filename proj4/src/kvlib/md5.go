package kvlib

import(
  "crypto/md5"
  "encoding/hex"
)

func str_MD5(text string) string {
   hash := md5.Sum([]byte(text))
   return hex.EncodeToString(hash[:])
}
func MD5(text []byte) string {
   hash := md5.Sum(text)
   return hex.EncodeToString(hash[:])
}
