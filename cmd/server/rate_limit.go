package main
import("net/http";"strconv";"sync";"time";"github.com/gin-gonic/gin")
type ipWindow struct{count int;expires time.Time}
type ipLimiter struct{mu sync.Mutex;windows map[string]ipWindow;max int;duration time.Duration}
func newIPLimiter(max int,duration time.Duration)*ipLimiter{return &ipLimiter{windows:make(map[string]ipWindow),max:max,duration:duration}}
func(l *ipLimiter)allow(ip string,now time.Time)(bool,int){
 l.mu.Lock();defer l.mu.Unlock()
 if len(l.windows)>10000{for key,w:=range l.windows{if !now.Before(w.expires){delete(l.windows,key)}}}
 w,ok:=l.windows[ip];if !ok||!now.Before(w.expires){w=ipWindow{expires:now.Add(l.duration)}}
 if w.count>=l.max{retry:=int(w.expires.Sub(now).Seconds())+1;if retry<1{retry=1};return false,retry}
 w.count++;l.windows[ip]=w;return true,0
}
func(l *ipLimiter)middleware()gin.HandlerFunc{return func(c *gin.Context){allowed,retry:=l.allow(c.ClientIP(),time.Now());if !allowed{c.Header("Retry-After",strconv.Itoa(retry));c.AbortWithStatusJSON(http.StatusTooManyRequests,gin.H{"error":"Too many registration attempts. Please try again later."});return};c.Next()}}
