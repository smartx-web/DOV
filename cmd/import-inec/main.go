// Command import-inec imports the Benue South subset of the independent
// nigeria-inec-geo snapshot. It is NOT an official INEC API integration.
// Usage: go run ./cmd/import-inec -dir ./data
// Directory must contain lgas.json, wards.json, polling-units.csv from
// https://github.com/saidiadegoke/nigeria-inec-geo/tree/main/data
package main

import (
 "context"
 "encoding/csv"
 "encoding/json"
 "flag"
 "fmt"
 "io"
 "os"
 "path/filepath"
 "strings"
 "time"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
)

const state = "07"
var allowed = map[string]bool{"ADO":true,"AGATU":true,"APA":true,"OBI":true,"OGBADIBO":true,"OHIMINI":true,"OJU":true,"OKPOKWU":true,"OTUKPO":true}
type Area struct { Code string `json:"code"`; Name string `json:"name"`; StateCode string `json:"state_code"`; LGACode string `json:"lga_code"` }
func readAreas(path string)([]Area,error){f,e:=os.Open(path);if e!=nil{return nil,e};defer f.Close();var a []Area;e=json.NewDecoder(f).Decode(&a);return a,e}
func main(){
 dir:=flag.String("dir","data","snapshot data directory");dry:=flag.Bool("dry-run",false,"validate and count only");flag.Parse()
 if err:=run(*dir,*dry);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
}
func run(dir string,dry bool)error{
 lgas,e:=readAreas(filepath.Join(dir,"lgas.json"));if e!=nil{return e}
 wards,e:=readAreas(filepath.Join(dir,"wards.json"));if e!=nil{return e}
 selected:=map[string]string{};for _,l:=range lgas{if l.StateCode==state&&allowed[strings.ToUpper(strings.TrimSpace(l.Name))]{selected[l.Code]=l.Name}}
 if len(selected)!=9{return fmt.Errorf("expected 9 Benue South LGAs, found %d",len(selected))}
 type W struct{code,name string};wm:=map[string]W{}
 for _,w:=range wards{if w.StateCode==state&&selected[w.LGACode]!=""{key:=w.LGACode+"/"+w.Code;if _,ok:=wm[key];ok{return fmt.Errorf("duplicate ward %s",key)};wm[key]=W{w.Code,w.Name}}}
 if len(wm)==0{return fmt.Errorf("no wards found")}
 f,e:=os.Open(filepath.Join(dir,"polling-units.csv"));if e!=nil{return e};defer f.Close()
 reader:=csv.NewReader(f);reader.FieldsPerRecord=-1;head,e:=reader.Read();if e!=nil{return e};idx:=map[string]int{};for i,h:=range head{idx[strings.TrimSpace(h)]=i}
 fields:=[]string{"state_code","lga_code","ward_code","pu_code","pu_name","full_code"}
 for _,k:=range fields{if _,ok:=idx[k];!ok{return fmt.Errorf("missing CSV column %s",k)}}
 var tx pgx.Tx;var db *pgxpool.Pool;ctx,cancel:=context.WithTimeout(context.Background(),3*time.Minute);defer cancel()
 if !dry{url:=os.Getenv("DATABASE_URL");if url==""{return fmt.Errorf("DATABASE_URL required")};db,e=pgxpool.New(ctx,url);if e!=nil{return e};defer db.Close();tx,e=db.Begin(ctx);if e!=nil{return e};defer tx.Rollback(ctx);for _,ddl:=range []string{"CREATE TABLE IF NOT EXISTS lgas(id BIGSERIAL PRIMARY KEY,name TEXT UNIQUE NOT NULL)","CREATE TABLE IF NOT EXISTS wards(id BIGSERIAL PRIMARY KEY,lga_id BIGINT NOT NULL REFERENCES lgas(id),name TEXT NOT NULL,UNIQUE(lga_id,name))","CREATE TABLE IF NOT EXISTS polling_units(id BIGSERIAL PRIMARY KEY,ward_id BIGINT NOT NULL REFERENCES wards(id),code TEXT NOT NULL,name TEXT NOT NULL,UNIQUE(ward_id,code))"}{if _,e=tx.Exec(ctx,ddl);e!=nil{return e}}}
 ids:=map[string]int64{};for code,name:=range selected{
 if !dry{var id int64;e=tx.QueryRow(ctx,"INSERT INTO lgas(name) VALUES($1) ON CONFLICT(name) DO UPDATE SET name=EXCLUDED.name RETURNING id",strings.Title(strings.ToLower(name))).Scan(&id);if e!=nil{return e};ids[code]=id}
 }
 wardIDs:=map[string]int64{};for key,w:=range wm{parts:=strings.Split(key,"/");if !dry{var id int64;e=tx.QueryRow(ctx,"INSERT INTO wards(lga_id,name) VALUES($1,$2) ON CONFLICT(lga_id,name) DO UPDATE SET name=EXCLUDED.name RETURNING id",ids[parts[0]],w.name).Scan(&id);if e!=nil{return e};wardIDs[key]=id}}
 seen:=map[string]bool{};count:=0
 for{row,err:=reader.Read();if err==io.EOF{break};if err!=nil{return err};get:=func(k string)string{if idx[k]>=len(row){return ""};return strings.TrimSpace(row[idx[k]])};if get("state_code")!=state||selected[get("lga_code")]==""{continue}
 wk:=get("lga_code")+"/"+get("ward_code");if _,ok:=wm[wk];!ok{return fmt.Errorf("polling unit has unknown ward %s",wk)}
 code:=get("pu_code");full:=get("full_code");name:=get("pu_name")
 if code==""||name==""||full==""{return fmt.Errorf("incomplete polling unit in %s",wk)}
 key:=wk+"/"+code;if seen[key]{return fmt.Errorf("duplicate polling unit %s",key)};seen[key]=true
 if !dry{_,e=tx.Exec(ctx,"INSERT INTO polling_units(ward_id,code,name) VALUES($1,$2,$3) ON CONFLICT(ward_id,code) DO UPDATE SET name=EXCLUDED.name",wardIDs[wk],code,name);if e!=nil{return e}}
 count++
 }
 if count==0{return fmt.Errorf("no Benue South polling units found")}
 fmt.Printf("Snapshot validated: %d LGAs, %d wards, %d polling units (dry run: %t)\n",len(selected),len(wm),count,dry)
 if !dry{return tx.Commit(ctx)};return nil
}
