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
)

func (receiver *Receiver) Like(entity federation.EntityLike) {
  var like models.Like
  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  revel.AppLog.Debug("Found a like entity", "entity", entity)

  err = like.Cast(&entity)
  if err != nil {
    // try to recover entity
    recovery := Recovery{models.ShareablePost, entity.ParentGuid}
    recovery.Run()

    err = like.Cast(&entity)
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
  }

  _, _, local := like.ParentPostUser()
  // if parent post is local we have
  // to relay the entity to all recipiens
  if local {
    // store order for later use
    order := models.SignatureOrder{
      Order: receiver.Entity.SignatureOrder,
    }
    err = order.CreateOrFind()
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    like.Signature.SignatureOrderID = order.ID
  }

  err = db.Create(&like).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  if local {
    dispatcher := Dispatcher{
      Model: like,
      Message: entity,
      Relay: true,
    }
    // relay the entity
    go dispatcher.Run()
  }
}
