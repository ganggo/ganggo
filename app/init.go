package app
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
  "io/ioutil"
  "encoding/json"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/ganggo/app/jobs"
  run "github.com/revel/modules/jobs/app/jobs"
  federation "git.feneas.org/ganggo/federation"
  "net/http"
  "strings"
  "fmt"
  "time"
  "github.com/getsentry/raven-go"
)

// will be set on compile time
var AppVersion = "v0"

func InitDB() {
  revel.Config.SetSection("ganggo")
  driver, found := revel.Config.String("db.driver")
  if !found {
    panic("Datbase config missing")
  }
  user, found := revel.Config.String("db.user")
  if !found {
    panic("Datbase config missing")
  }
  password, found := revel.Config.String("db.password")
  if !found {
    panic("Datbase config missing")
  }
  database, found := revel.Config.String("db.database")
  if !found {
    panic("Datbase config missing")
  }
  host, found := revel.Config.String("db.host")
  if !found {
    panic("Datbase config missing")
  }
  dsn, found := revel.Config.String("db.dsn")
  if !found {
    panic("Datbase config missing")
  }

  models.DB.Driver = driver
  models.DB.Url = fmt.Sprintf(dsn, user, password, host, database)
  models.InitDB()
}

func InitSocialRelay() {
  // register the pod at the-federation.info
  // this is required for using the social-relay
  revel.Config.SetSection("ganggo")
  subscribe := revel.Config.BoolDefault("relay.subscribe", false)
  address, found := revel.Config.String("address")
  if !revel.DevMode && subscribe && found {
    result := struct{Error string `json:"error"`}{}
    endpoint := fmt.Sprintf(revel.Config.StringDefault(
      "relay.endpoint", "https://the-federation.info/register/%s",
    ), address)

    err := federation.FetchJson("GET", endpoint, nil, &result)
    if err != nil {
      revel.AppLog.Error("InitSocialRelay failed",
        "result", result, "err", err)
    } else {
      revel.AppLog.Info("InitSocialRelay registration", "result", result)
    }
  } else {
    revel.AppLog.Info("InitSocialRelay skipped",
      "devMode", revel.DevMode, "subscribe", subscribe)
  }
}

func init() {
  // Filters is the default set of global filters.
  revel.Filters = []revel.Filter{
    revel.PanicFilter,             // Recover from panics and display an error page instead.
    revel.RouterFilter,            // Use the routing table to select the right Action
    revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
    revel.ParamsFilter,            // Parse parameters into Controller.Params.
    JsonParamsFilter,
    revel.SessionFilter,           // Restore and write the session cookie.
    revel.FlashFilter,             // Restore and write the flash cookie.
    revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
    revel.I18nFilter,              // Resolve the requested language
    HeaderFilter,                  // Add some security based headers
    revel.InterceptorFilter,       // Run interceptors around the action.
    revel.CompressFilter,          // Compress the result.
    revel.ActionInvoker,           // Invoke the action.
  }

  // register startup functions with OnAppStart
  revel.OnAppStart(InitDB)
  revel.OnAppStart(InitSocialRelay)

  revel.OnAppStart(func() {
    revel.Config.SetSection("ganggo")

    // configure the federation library
    host := revel.Config.StringDefault("proto", "http://") +
      revel.Config.StringDefault("address", "localhost:9000")
    apiVersion := revel.Config.StringDefault("api.version", "v0")
    federation.SetConfig(federation.Config{
      Host: host, ApiVersion: apiVersion,
      GuidURLFormat: host + "/posts/%s",
      ApURLFormat: host + "/api/" + apiVersion + "/ap/%s",
    })

    // set custom logger options
    federation.SetLogger(helpers.AppLogWrapper{
      Name: "federation",
    })

    // if sentry credentials exists
    // send reports to upstream
    sentryDSN, found := revel.Config.String("sentry.DSN")
    if found {
      raven.SetDSN(sentryDSN)
      raven.SetRelease(AppVersion)
      revel.RootLog.SetHandler(helpers.SentryLogHandler{})
    }
  })

  // register jobs running on an interval
  revel.OnAppStart(func() {
    run.Every(24*time.Hour, jobs.Session{})

    revel.Config.SetSection("ganggo")
    if revel.Config.BoolDefault("telegram.enabled", false) {
      host := revel.Config.StringDefault("proto", "http://") +
        revel.Config.StringDefault("address", "localhost:9000")
      token := revel.Config.StringDefault("telegram.token", "")
      run.Now(jobs.TelegramWebhook{Token: token, Url: host})
    }
  })
}

// TODO turn this into revel.HeaderFilter
// should probably also have a filter for CSRF
// not sure if it can go in the same filter or not
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
  // Add some common security headers
  c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
  c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
  c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

  fc[0](c, fc[1:]) // Execute the next filter stage.
}

var JsonParamsFilter = func(c *revel.Controller, fc []revel.Filter) {
  if strings.Contains(c.Request.ContentType, "application/json") {
    data := map[string]string{}
    request := c.Request.In.GetRaw().(*http.Request)
    content, _ := ioutil.ReadAll(request.Body)
    json.Unmarshal(content, &data)
    for k, v := range data {
      revel.TRACE.Println("application/json", k, v)
      c.Params.Values.Set(k, v)
    }
  }
  fc[0](c, fc[1:])
}
