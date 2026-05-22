package main
import ("encoding/json";"fmt";"net/url";"os";"github.com/wepie/qianji")
func main(){
 data,_:=os.ReadFile(os.Getenv("HOME")+"/.qianji_token.json")
 var m map[string]string
 json.Unmarshal(data,&m)
 c:=qianji.NewClient("")
 s:=&qianji.Session{Client:c,Token:m["token"],UserID:m["user_id"]}
 // 发送纯删除请求
 p:=map[string]interface{}{"bills":map[string]interface{}{"dellist":[]int64{1779436792803427}}}
 j,_:=json.Marshal(p)
 params:=url.Values{}
 params.Set("uid",s.UserID);params.Set("fr",s.UserID);params.Set("v",string(j))
 raw,_:=s.Client.DoPostRaw("bill","syncall",params,s.Token)
 fmt.Println(string(raw))
}
