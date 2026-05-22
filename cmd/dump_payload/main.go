package main
import ("encoding/json";"fmt";"os";"github.com/wepie/qianji")
func main(){
 data,_:=os.ReadFile(os.Getenv("HOME")+"/.qianji_token.json")
 var m map[string]string
 json.Unmarshal(data,&m)
 c:=qianji.NewClient("")
 s:=&qianji.Session{Client:c,Token:m["token"],UserID:m["user_id"]}
 b:=qianji.NewBill(0,1,"dump测试")
 payload:=map[string]interface{}{
  "bills":map[string]interface{}{
   "changelist":[]map[string]interface{}{{
    "id":b.ID,"userid":s.UserID,"time":b.TimeInSec,"type":0,
    "remark":"dump测试","money":1.0,"status":2,"cateid":89693200,
    "updatetime":b.TimeInSec,"createtime":b.TimeInSec,
    "platform":0,"assetid":-1,"fromid":-1,"targetid":-1,
    "bookid":-1,"packid":-1,"images":[]string{},
   }}}}
 j,_:=json.Marshal(payload)
 fmt.Println(string(j))
}
