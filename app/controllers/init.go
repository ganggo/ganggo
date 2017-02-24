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
  "github.com/ganggo/ganggo/app/models"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

var _arrange bool

func init() {
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Stream{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Search{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Profile{})

  revel.TemplateFuncs["isLoggedIn"] = func(token interface {}) bool {
    switch token.(type) {
    case string:
      return true
    }
    return false
  }
  revel.TemplateFuncs["person"] = func(id uint) (person models.Person) {
    db, err := gorm.Open(models.DB.Driver, models.DB.Url)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    defer db.Close()

    err = db.First(&person, id).Error
    if err != nil {
      revel.ERROR.Println(err, id)
      return
    }

    err = db.Where("person_id = ?", person.ID).First(&person.Profile).Error
    if err != nil {
      revel.ERROR.Println(err, person)
      return
    }
    return
  }
  revel.TemplateFuncs["swapElement"] = func() bool {
    return _arrange
  }
  revel.TemplateFuncs["toggleSwapElement"] = func() bool {
    if _arrange {
      _arrange = false
    } else {
      _arrange = true
    }
    return _arrange
  }
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
  c.RenderArgs["TOKEN"] = c.Session["TOKEN"]
  return nil
}
