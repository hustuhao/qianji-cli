package main
import ("encoding/json";"fmt";"net/url";"os";"time";"math/rand";"github.com/wepie/qianji")
func main(){
 data,_:=os.ReadFile(os.Getenv("HOME")+"/.qianji_token.json")
 var m map[string]string
 json.Unmarshal(data,&m)
 c:=qianji.NewClient("")
 s:=&qianji.Session{Client:c,Token:m["token"],UserID:m["user_id"]}
 now:=time.Now().Unix()
 billid:=time.Now().UnixMilli()*1000+rand.Int63n(1000)
 // 完全模仿手机 Bill.toSyncJson() 的输出
 payload:=map[string]interface{}{
  "bills":map[string]interface{}{
   "changelist":[]map[string]interface{}{{
    "assetid":-1,"bookid":-1,"cateid":89693200,
    "createtime":now,"descinfo":"","extra":"",
    "fromid":-1,"id":billid,"images":[]string{},
    "money":1.0,"packid":-1,"platform":0,
    "remark":"精确模仿","status":2,"targetid":-1,
    "time":now,"type":0,"updatetime":now,
    "userid":s.UserID,"username":"",
   }}}}
 j,_:=json.Marshal(payload)
 fmt.Println("PAYLOAD:",string(j))
 params:=url.Values{}
 params.Set("uid",s.UserID);params.Set("fr",s.UserID);params.Set("v",string(j))
 raw,_:=s.Client.DoPostRaw("bill","syncall",params,s.Token)
 fmt.Println("RESPONSE:",string(raw))
}
