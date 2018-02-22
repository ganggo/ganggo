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
)

type Hcard struct {
  *revel.Controller
}

func (c Hcard) User(guid string) revel.Result {
  var person models.Person

  db, err := models.OpenDatabase()
  if err != nil {
    c.Log.Error("Cannot open database", "error", err)
    return c.RenderError(err)
  }
  defer db.Close()

  err = db.Where("guid = ?", guid).First(&person).Error
  if err != nil {
    c.Log.Error("Person not found", "error", err)
    return c.NotFound(err.Error())
  }

  var profile models.Profile
  err = db.Where("person_id = ?", person.ID).First(&profile).Error
  if err != nil {
    c.Log.Error("Profile not found", "error", err)
    return c.NotFound(err.Error())
  }

  revel.Config.SetSection("ganggo")
  proto := revel.Config.StringDefault("proto", "http://")
  address := revel.Config.StringDefault("address", "localhost")

  c.ViewArgs["profile"] = profile
  c.ViewArgs["person"] = person
  c.ViewArgs["proto"] = proto
  c.ViewArgs["address"] = address

  return c.Render()
}
