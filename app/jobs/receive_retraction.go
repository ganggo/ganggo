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
)

func (r *Receiver) Retraction(entity federation.MessageRetract) {
  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  // relay retraction before deleting database entries!
  (&Dispatcher{Message: entity}).Run()

  var dbModel interface{}
  switch entity.ParentType() {
  case federation.Reshare:
    fallthrough
  case federation.StatusMessage:
    var post models.Post
    err := post.FindByGuid(entity.ParentGuid())
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    dbModel = &post
  case federation.Comment:
    var comment models.Comment
    err = comment.FindByGuid(entity.ParentGuid())
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    dbModel = &comment
  case federation.Like:
    var like models.Like
    err = like.FindByGuid(entity.ParentGuid())
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    dbModel = &like
  case federation.Unknown:
    // NOTE in case of mastodon we do not know if it was
    // a post or comment since we only receive a tombstone
    // activity with the id or guid as reference
    var (
      post models.Post
      comment models.Comment
    )
    postErr := post.FindByGuid(entity.ParentGuid())
    commentErr := comment.FindByGuid(entity.ParentGuid())
    if postErr == nil {
      dbModel = &post
    } else if commentErr == nil {
      dbModel = &comment
    }
  default:
    revel.AppLog.Error(
      "Unkown TargetType in Dispatcher", "retraction", entity)
  }

  if dbModel != nil {
    err = db.Delete(dbModel).Error
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
  } else {
    revel.AppLog.Error("DB model is nil!", "retraction", entity)
  }
}
