package main
import ("encoding/json";"fmt";"os";"github.com/wepie/qianji")
func main(){
 data,_:=os.ReadFile(os.Getenv("HOME")+"/.qianji_token.json")
 var m map[string]string
 json.Unmarshal(data,&m)
 c:=qianji.NewClient("")
 s:=&qianji.Session{Client:c,Token:m["token"],UserID:m["user_id"]}
 b:=qianji.NewBill(0,1,"catetest")
 b=b.WithCategory(89693183)
 ids,err:=s.SyncBills([]qianji.Bill{b},nil)
 fmt.Printf("ids=%v err=%v\n",ids,err)
 // also dump what the syncall actually sent
 type syncBill struct{BillID int64 `json:"id"`;CateID int64 `json:"cateid"`;Remark string `json:"remark"`;Money float64 `json:"money"`;Status int `json:"status"`;UserID string `json:"userid"`;TimeInSec int64 `json:"time"`;Type int `json:"type"`;Up int64 `json:"updatetime"`;Cr int64 `json:"createtime"`;Pl int `json:"platform"`;As int64 `json:"assetid"`;Fr int64 `json:"fromid"`;Ta int64 `json:"targetid"`;Ex string `json:"extra"`;De string `json:"descinfo"`;Bo int64 `json:"bookid"`;Un string `json:"username"`;Pa int64 `json:"packid"`;Im []string `json:"images"`}
 sb:=syncBill{BillID:b.ID,CateID:b.CateID,Remark:b.Remark,Money:b.Money,Status:2,UserID:s.UserID,TimeInSec:b.TimeInSec,Type:0,Up:b.UpdateTime,Cr:b.CreateTime,Pl:0,As:-1,Fr:-1,Ta:-1,Ex:"",De:"",Bo:-1,Un:"",Pa:-1,Im:[]string{}}
 j,_:=json.Marshal(sb)
 fmt.Println("syncBill JSON:",string(j))
}
