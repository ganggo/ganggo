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
  "time"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/models"
  federation "git.feneas.org/ganggo/federation"
  run "github.com/revel/modules/jobs/app/jobs"
)

func (receiver *Receiver) Comment(entity federation.MessageComment) {
  var (
    post models.Post
    person models.Person
  )

  var postID uint
  err := post.FindByGuid(entity.Parent())
  if err == nil {
    postID = post.ID
  } else {
    // this can happen with e.g. mastodon which always replies to
    // the previous comment entity not the original post
    var prevComment models.Comment
    err = prevComment.FindByGuid(entity.Parent())
    if err != nil {
      revel.AppLog.Error("Receiver Comment", err.Error(), err)
      return
    }
    postID = prevComment.ShareableID
  }

  err = person.FindByAuthor(entity.Author())
  if err != nil {
    revel.AppLog.Error("Receiver Comment", err.Error(), err)
    return
  }

  createdAt, err := entity.CreatedAt().Time()
  if err != nil {
    createdAt = time.Now()
  }

  comment := models.Comment{
    CreatedAt: createdAt,
    Text: entity.Text(),
    ShareableID: postID,
    PersonID: person.ID,
    Guid: entity.Guid(),
    ShareableType: models.ShareablePost,
    Protocol: entity.Type().Proto,
  }

  _, _, local := comment.ParentPostUser()
  // if parent post is local we have
  // to relay the comment to all recipiens
  if local {
    order := models.SignatureOrder{Order: entity.SignatureOrder()}
    err = order.CreateOrFind()
    if err != nil {
      revel.AppLog.Error("Receiver Comment", err.Error(), err)
      return
    }
    comment.Signature.SignatureOrderID = order.ID
  }
  comment.Signature.AuthorSignature = entity.Signature()

  err = comment.Create()
  if err != nil {
    // XXX RECOVERY
    //// try to recover entity
    //recovery := Recovery{models.ShareablePost, entity.ParentGuid}
    //recovery.Run()

    //err = comment.Cast(&entity)
    //if err != nil {
      revel.AppLog.Error("Receiver Comment", err.Error(), err)
      return
    //}
  }

  if local {
    run.Now(Dispatcher{Message: entity})
  }
}
