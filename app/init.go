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
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/views"
  "github.com/shaoshing/train"
  "strings"
  "fmt"
)

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

func init() {
  // Filters is the default set of global filters.
  revel.Filters = []revel.Filter{
    AssetsFilter,
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

  train.Config.AssetsPath = "app/assets"
  train.Config.SASS.DebugInfo = false
  train.Config.SASS.LineNumbers = false
  train.Config.Verbose = false
  train.Config.BundleAssets = true

  // assets
  train.ConfigureHttpHandler(nil)
  revel.TemplateFuncs["javascript_include_tag"] = train.JavascriptTag
  revel.TemplateFuncs["stylesheet_link_tag"] = train.StylesheetTag

  // append custom template functions to revel
  for key, val := range views.TemplateFuncs {
    revel.TemplateFuncs[key] = val
  }
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

var AssetsFilter = func(c *revel.Controller, fc []revel.Filter) {
  path := c.Request.URL.Path
  if strings.HasPrefix(path, "/assets") {
    train.ServeRequest(c.Response.Out, c.Request.Request)
  } else {
    fc[0](c, fc[1:])
  }
}

var JsonParamsFilter = func(c *revel.Controller, fc []revel.Filter) {
  if strings.Contains(c.Request.ContentType, "application/json") {
    data := map[string]string{}
    content, _ := ioutil.ReadAll(c.Request.Body)
    json.Unmarshal(content, &data)
    for k, v := range data {
      revel.TRACE.Println("application/json", k, v)
      c.Params.Values.Set(k, v)
    }
  }
  fc[0](c, fc[1:])
}
