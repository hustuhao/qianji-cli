package main
import ("encoding/json";"fmt";"net/url";"os";"github.com/wepie/qianji")
func main(){
 data,_:=os.ReadFile(os.Getenv("HOME")+"/.qianji_token.json")
 var m map[string]string
 json.Unmarshal(data,&m)
 c:=qianji.NewClient("")
 s:=&qianji.Session{Client:c,Token:m["token"],UserID:m["user_id"]}

 // 用不同 devid 拉取
 devids:=[]string{"","deadbeefdeadbeefdeadbeefdeadbeef"}
 for _,did:=range devids{
  params:=url.Values{}
  params.Set("uid",s.UserID);params.Set("fr",s.UserID);params.Set("bookid","-1");params.Set("pageoffset","0");params.Set("pagesign","")
  var raw []byte
  if did==""{
   raw,_=s.Client.DoPostRaw("syncv2","pull",params,s.Token)
  }else{
   raw,_=s.Client.DoPostCustom("syncv2","pull",params,s.Token,did)
  }
  var resp struct{Data struct{Count int `json:"count"`} `json:"data"`}
  json.Unmarshal(raw,&resp)
  fmt.Printf("devid=%s count=%d\n",did,resp.Data.Count)
 }
}
