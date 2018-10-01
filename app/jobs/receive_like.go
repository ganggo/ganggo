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
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/federation"
  run "github.com/revel/modules/jobs/app/jobs"
)

func (receiver *Receiver) Like(entity federation.MessageLike) {
  var (
    post models.Post
    person models.Person
  )

  err := post.FindByGuid(entity.Parent())
  if err != nil {
    revel.AppLog.Error("Receiver Like", err.Error(), err)
    return
  }

  err = person.FindByAuthor(entity.Author())
  if err != nil {
    revel.AppLog.Error("Receiver Like", err.Error(), err)
    return
  }

  like := models.Like{
    Positive: entity.Positive(),
    ShareableID: post.ID,
    PersonID: person.ID,
    Guid: entity.Guid(),
    ShareableType: models.ShareablePost,
    Protocol: entity.Type().Proto,
  }
  _, _, local := like.ParentPostUser()
  // if parent post is local we have
  // to relay the entity to all recipiens
  if local {
    // store order for later use
    order := models.SignatureOrder{Order: entity.SignatureOrder()}
    err = order.CreateOrFind()
    if err != nil {
      revel.AppLog.Error("Receiver Like", err.Error(), err)
      return
    }
    like.Signature.SignatureOrderID = order.ID
  }
  like.Signature.AuthorSignature = entity.Signature()

  err = like.Create()
  if err != nil {
    // try to recover entity
    // XXX recovery
    //recovery := Recovery{models.ShareablePost, entity.ParentGuid}
    //recovery.Run()

    //err = like.Cast(&entity)
    //if err != nil {
      revel.AppLog.Error("Receiver Like", err.Error(), err)
      return
    //}
  }

  if local {
    run.Now(Dispatcher{Message: entity})
  }
}
