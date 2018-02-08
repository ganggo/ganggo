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
)

func (r *Receiver) Retraction(retraction federation.EntityRetraction) {
  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  // NOTE relay to other hosts if we own this entity
  // should be done before we start deleting db records
  dispatcher := Dispatcher{Message: retraction}
  dispatcher.Run()

  if strings.EqualFold(retraction.TargetType, models.ShareablePost) {
    var post models.Post
    err = post.FindByGuid(retraction.TargetGuid)
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    err = db.Delete(&post).Error
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
  } else if strings.EqualFold(retraction.TargetType, models.ShareableComment) {
    var comment models.Comment
    err = comment.FindByGuid(retraction.TargetGuid)
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    err = db.Delete(&comment).Error
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
  } else if strings.EqualFold(retraction.TargetType, models.ShareableLike) {
    var like models.Like
    err = like.FindByGuid(retraction.TargetGuid)
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    err = db.Delete(&like).Error
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
  } else {
    revel.AppLog.Error("Unkown TargetType in Dispatcher", "retraction", retraction)
    return
  }
}
