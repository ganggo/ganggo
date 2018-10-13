package helpers
//
// GangGo Application Server
// Copyright (C) 2017 Lukas Matt <lukas@zauberstuhl.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

import (
  "fmt"
  "github.com/revel/revel"
  federation "git.feneas.org/ganggo/federation"
  "github.com/revel/log15"
  "github.com/getsentry/raven-go"
  "errors"
)

type AppLogWrapper struct {
  Name string
}

// SentryLogHandler will intercept logging and send all errors
// with a log level greater then info to the specified DSN
type SentryLogHandler struct {}

func (handler SentryLogHandler) Log(record *log15.Record) error {
  if record.Lvl == log15.LvlError || record.Lvl == log15.LvlCrit {
    var errs []error
    // search for errors in the logger context
    for _, ctx := range record.Ctx {
      if err, ok := ctx.(error); ok {
        errs = append(errs, err)
      }
    }

    if len(errs) == 0 {
      // there was an error/crit event but no error type
      // was found lets create one and send it to sentry
      var errMsg string
      for _, ctx := range record.Ctx {
        if msg, ok := ctx.(string); ok {
          errMsg = errMsg + " " + msg
        }
      }
      errs = append(errs, errors.New(errMsg))
    }

    // send asynchronously to the sentry endpoint
    for _, err := range errs {
      raven.CaptureError(err, nil)
    }
  }
  return revel.GetRootLogHandler().Log(record)
}

func (wrap AppLogWrapper) Println(v ...interface{}) { wrap.Print(v) }

func (wrap AppLogWrapper) Print(v ...interface{}) {
  if len(v) <= 0 {
    return
  }

  if wrap.Name == "gorm" {
    logType := v[0].(string)
    if logType == "log" && len(v) > 2 {
      path := v[1].(string)
      revel.AppLog.Error(path, "error", v[2])
      return
    } else if logType == "sql" && len(v) > 5 {
      path := v[1].(string)
      revel.AppLog.Debug(path, "time", v[2],
        "query", v[3], "params", v[4], "rows", v[5])
      return
    }
  } else if wrap.Name == "federation" {
    var logType string
    switch log := v[0].(type) {
    case []interface{}:
      logType = log[0].(string)
    case interface{}:
      logType = log.(string)
    }

    if logType == federation.LOG_C_RED {
      revel.AppLog.Error(fmt.Sprintf("%+v", v))
    } else if logType == federation.LOG_C_YELLOW {
      revel.AppLog.Warn(fmt.Sprintf("%+v", v))
    } else {
      revel.AppLog.Debug(fmt.Sprintf("%+v", v))
    }
    return
  }
  revel.AppLog.Debug(fmt.Sprintf("%+v", v))
}
