# rlog
Leveled execution logs for Go, MongoDb logger for GoLang
========================================================
# Start rlog for mongodb
var log *rlog.Logger
  func main() {
        log, err = rlog.NewLoggerDial(rlog.LoggerTypeMongoDb, "", "localhost:27017/dbname", "")
        if err != nil {
                panic(err)
        }
  }
# Use the log
  log.Err(err.Error())
# or
  log.Fatal(err.Error())
# or
  log.Info("This is rlog")
# Use std functions of log
  log.Println(err)




