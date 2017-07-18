package jobs
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
  "strings"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  federation "gopkg.in/ganggo/federation.v0"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

func (r *Receiver) Retraction(entity federation.EntityRetraction) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  switch {
  case strings.EqualFold("post", entity.TargetType):
    var post models.Post
    err := db.Where("guid = ?", entity.TargetGuid).First(&post).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
    err = db.Delete(&post).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
  case strings.EqualFold("like", entity.TargetType):
    var like models.Like
    err := db.Where("guid = ?", entity.TargetGuid).First(&like).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
    err = db.Delete(&like).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
  case strings.EqualFold("comment", entity.TargetType):
    var comment models.Comment
    err := db.Where("guid = ?", entity.TargetGuid).First(&comment).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
    err = db.Delete(&comment).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
  default:
    revel.ERROR.Println("Unknown entity:", entity)
    return
  }
}
