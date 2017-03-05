package views
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
  "github.com/ganggo/ganggo/app/helpers"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

var _templateFuncs = map[string]interface{}{
  "isLoggedIn": func(token interface {}) bool {
    switch token.(type) {
    case string:
      return true
    }
    return false
  },
  "person": func(id uint) (person models.Person) {
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
  },
  "HostFromHandle": func(handle string) (host string) {
    _, host, err := helpers.ParseDiasporaHandle(handle)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    return
  },
}

func init() {
  // append custom template functions to revel
  for key, val := range _templateFuncs {
    revel.TemplateFuncs[key] = val
  }
}
