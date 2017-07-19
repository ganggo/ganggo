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
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  federation "gopkg.in/ganggo/federation.v0"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

func (r *Receiver) Like(entity federation.EntityLike) {
  var like models.Like
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  revel.TRACE.Println("Found a like entity", entity)

  err = like.Cast(&entity)
  if err != nil {
    revel.WARN.Println(err)
    return
  }

  user, local := like.ParentIsLocal()
  // if parent post is local we have
  // to relay the entity to all recipiens
  if local {
    revel.TRACE.Println("Parent is local! Relaying it..")

    sigOrder := models.SignatureOrder{
      Order: r.Entity.SignatureOrder,
    }
    if err = sigOrder.CreateOrFind(); err != nil {
      revel.ERROR.Println(err)
      return
    }
    like.Signature.SignatureOrderID = sigOrder.ID

    var visibilities models.AspectVisibilities
    err = db.Where(
      "shareable_id = ? and shareable_type = ?",
      like.TargetID,
      like.TargetType,
    ).Find(&visibilities).Error
    if err != nil {
      revel.ERROR.Println(err)
      return
    }

    if len(visibilities) == 0 {
      revel.TRACE.Println(".. relaying it publicly!")
      go r.RelayPublic(user)
    } else {
      revel.TRACE.Println(".. relaying it privately!")
      go r.RelayPrivate(user, visibilities)
    }
  }

  err = db.Create(&like).Error
  if err != nil {
    revel.WARN.Println(err)
    return
  }
}
