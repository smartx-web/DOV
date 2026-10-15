package main
import("testing";"time")
func TestIPLimiter(t *testing.T){l:=newIPLimiter(2,time.Minute);now:=time.Date(2026,10,15,12,0,0,0,time.UTC);for i:=0;i<2;i++{ok,_:=l.allow("one",now);if !ok{t.Fatal("request blocked too soon")}};if ok,retry:=l.allow("one",now);ok||retry<=0{t.Fatal("limit not enforced")};if ok,_:=l.allow("two",now);!ok{t.Fatal("different client blocked")};if ok,_:=l.allow("one",now.Add(time.Minute));!ok{t.Fatal("window did not reset")}}
