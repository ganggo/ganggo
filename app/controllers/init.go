package controllers
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
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  api "gopkg.in/ganggo/api.v0/app/controllers"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

func init() {
  // redirect if logged-in
  revel.InterceptFunc(redirectIfLoggedIn, revel.BEFORE, &App{})
  // requires login
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Stream{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Post{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Search{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Profile{})
  // API
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &api.ApiComment{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &api.ApiLike{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &api.ApiPost{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &api.ApiPeople{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &api.ApiProfile{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &api.ApiAspect{})
}

func redirectIfLoggedIn(c *revel.Controller) revel.Result {
  result := requiresLogin(c)
  if result == nil {
    return c.Redirect(Stream.Index)
  }
  return nil
}

func requiresLogin(c *revel.Controller) revel.Result {
  var session models.Session

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return c.Render()
  }
  defer db.Close()

  err = db.Where("token = ?", c.Session["TOKEN"]).First(&session).Error
  if err != nil {
    c.Flash.Error("Please log in first")
    return c.Redirect(App.Index)
  }
  c.ViewArgs["TOKEN"] = c.Session["TOKEN"]
  return nil
}
