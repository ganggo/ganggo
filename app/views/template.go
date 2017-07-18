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
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  "github.com/jinzhu/gorm"
  "github.com/dchest/captcha"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

var TemplateFuncs = map[string]interface{}{
  // database types
  "IsReshare": func(a string) bool {
    return (a == models.Reshare)
  },
  "IsStatusMessage": func(a string) bool {
    return (a == models.StatusMessage)
  },
  "IsShareablePost": func(a string) bool {
    return (a == models.ShareablePost)
  },
  // session helper
  "IsLoggedIn": func(in interface {}) (user models.User) {
    switch token := in.(type) {
    case string:
      user, _ = models.GetCurrentUser(token)
    }
    return
  },
  "LikesByTargetID": func(id uint) []models.Like {
    return likes(id, true)
  },
  "DislikesByTargetID": func(id uint) []models.Like {
    return likes(id, false)
  },
  "PostByGuid": func(guid string) (post models.Post) {
    db, err := gorm.Open(models.DB.Driver, models.DB.Url)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    defer db.Close()

    err = db.Where("guid = ?", guid).First(&post).Error
    if err != nil {
      revel.ERROR.Println(err, guid)
      return
    }
    return
  },
  "PersonByID": func(id uint) (person models.Person) {
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
  // string parse helper
  "HostFromHandle": func(handle string) (host string) {
    _, host, err := helpers.ParseAuthor(handle)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    return
  },
  // captcha generator
  "CaptchaNew": func() string { return captcha.New() },
}

func likes(id uint, like bool) (likes []models.Like) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  defer db.Close()

  err = db.Where(
    `target_type = ?
      and target_id = ?
      and positive = ?`,
    models.ShareablePost, id, like,
  ).Find(&likes).Error
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  return
}
