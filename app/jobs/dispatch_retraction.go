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
  "encoding/xml"
  "github.com/revel/revel"
  "github.com/ganggo/ganggo/app/models"
  diaspora "github.com/ganggo/federation/diaspora"
  "strings"
)

func (dispatcher *Dispatcher) Retraction(retraction diaspora.EntityRetraction) {
  var (
    parentPost models.Post
    parentUser models.User
    ok bool
  )
  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  // NOTE I can only request retraction if I am the owner
  if strings.EqualFold(retraction.EntityTargetType, models.ShareablePost) {
    err = parentPost.FindByGuid(retraction.EntityTargetGuid)
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    err = parentUser.FindByID(parentPost.Person.UserID)
    if err != nil {
      revel.AppLog.Debug("We can only retract if we own the entity", err.Error())
      return
    }
  } else if strings.EqualFold(retraction.EntityTargetType, models.ShareableComment) {
    var comment models.Comment
    err = comment.FindByGuid(retraction.EntityTargetGuid)
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    if parentPost, parentUser, ok = comment.ParentPostUser(); !ok {
      revel.AppLog.Debug("We can only retract if we own the entity")
      return
    }
  } else if strings.EqualFold(retraction.EntityTargetType, models.ShareableLike) {
    var like models.Like
    err := like.FindByGuid(retraction.EntityTargetGuid)
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
    if parentPost, parentUser, ok = like.ParentPostUser(); !ok {
      revel.AppLog.Debug("We can only retract if we own the entity")
      return
    }
  } else {
    revel.AppLog.Error("Unkown TargetType in Dispatcher", "retraction", retraction)
    return
  }

  entityXml, err := xml.Marshal(retraction)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  dispatcher.Send(parentPost, parentUser, entityXml, 0)
}
