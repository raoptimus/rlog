# rlog
## Leveled execution logs for Go, MongoDb logger for GoLang
========================================================
### Start rlog for mongodb
```go
  var log *rlog.Logger
  func main() {
        log, err = rlog.NewLoggerDial(rlog.LoggerTypeMongoDb, "", "localhost:27017/dbname", "")
        if err != nil {
                panic(err)
        }
  }
```
### Use the log
```go
  log.Err(err.Error())
```
### or
```go
  log.Fatal(err.Error())
```
### or
```go
  log.Info("This is rlog")
```
### Use std functions of log
```go
  log.Println(err)
```



