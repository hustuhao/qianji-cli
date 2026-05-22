package main
import ("crypto/md5";"encoding/json";"fmt";"math/rand";"net/url";"os";"github.com/wepie/qianji")
func main(){
 data,_:=os.ReadFile(os.Getenv("HOME")+"/.qianji_token.json")
 var m map[string]string
 json.Unmarshal(data,&m)
 c:=qianji.NewClient("")
 s:=&qianji.Session{Client:c,Token:m["token"],UserID:m["user_id"]}
 uuid:=fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",rand.Uint32(),rand.Uint32()&0xffff,(rand.Uint32()&0x0fff)|0x4000,(rand.Uint32()&0x3fff)|0x8000,rand.Uint64()&0xffffffffffff)
 newDevID:=fmt.Sprintf("%x",md5.Sum([]byte(uuid)))
 params:=url.Values{}
 params.Set("uid",s.UserID)
 params.Set("bookid","-1")
 params.Set("pageoffset","0")
 params.Set("pagesign","")
 raw,err:=s.Client.DoPostCustom("syncv2","pull",params,s.Token,newDevID)
 if err!=nil{fmt.Println("err:",err);return}
 fmt.Println(string(raw))
}
