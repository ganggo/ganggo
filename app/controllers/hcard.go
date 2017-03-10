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
  "net/http"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Hcard struct {
  *revel.Controller
}

func (c Hcard) User() revel.Result {
  var (
    guid string
    person models.Person
  )
  c.Params.Bind(&guid, "guid")

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return c.Render()
  }
  defer db.Close()

  err = db.Where("guid = ?", guid).First(&person).Error
  if err != nil {
    c.Response.Status = http.StatusNotFound
    revel.WARN.Println(err)
    return c.Render()
  }

  var profile models.Profile
  err = db.Where("person_id = ?", person.ID).First(&profile).Error
  if err != nil {
    c.Response.Status = http.StatusNotFound
    revel.WARN.Println(err)
    return c.Render()
  }

  revel.Config.SetSection("ganggo")
  proto := revel.Config.StringDefault("proto", "http://")
  address := revel.Config.StringDefault("address", "localhost")

  c.RenderArgs["profile"] = profile
  c.RenderArgs["person"] = person
  c.RenderArgs["proto"] = proto
  c.RenderArgs["address"] = address

  return c.Render()
}
