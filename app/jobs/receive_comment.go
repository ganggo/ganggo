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

func (r *Receiver) Comment(entity federation.EntityComment) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  var comment models.Comment

  revel.TRACE.Println("Found a comment entity", entity)

  err = comment.Cast(&entity)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  user, local := comment.ParentIsLocal()
  // if parent post is local we have
  // to relay the comment to all recipiens
  if local {
    revel.TRACE.Println("Parent post is local! Relaying it..")

    sigOrder := models.SignatureOrder{
      Order: r.Entity.SignatureOrder,
    }
    if err = sigOrder.CreateOrFind(); err != nil {
      revel.ERROR.Println(err)
      return
    }
    comment.Signature.SignatureOrderID = sigOrder.ID

    var visibilities models.AspectVisibilities
    err = db.Where(
      "shareable_id = ? and shareable_type = ?",
      comment.ShareableID,
      comment.ShareableType,
    ).Find(&visibilities).Error
    if err != nil {
      revel.ERROR.Println(err)
      return
    }

    if len(visibilities) == 0 {
      revel.TRACE.Println(".. relaying it publicly!")
      go r.RelayPublicComment(user)
    } else {
      revel.TRACE.Println(".. relaying it privately!")
      go r.RelayPrivateComment(user, visibilities)
    }
  }

  err = db.Create(&comment).Error
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
}
