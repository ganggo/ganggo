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
  "github.com/ganggo/ganggo/app/models"
  federation "github.com/ganggo/federation"
)

func (receiver *Receiver) Reshare(entity federation.MessageReshare) {
  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  fetch := FetchAuthor{Author: entity.ParentAuthor()}
  fetch.Run()
  if fetch.Err != nil {
    revel.AppLog.Error(fetch.Err.Error())
    return
  }

  createdAt, err := entity.CreatedAt().Time()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  var post models.Post
  err = post.FindByGuid(entity.Parent())
  if err != nil {
    // XXX RECOVERY
    //// try to recover entity
    //recovery := Recovery{models.ShareablePost, entity.RootGuid}
    //recovery.Run()

    //err = post.FindByGuid(entity.RootGuid)
    //if err != nil {
      revel.AppLog.Error(err.Error())
      return
    //}
  }

  rootGuid := entity.Parent()
  reshare := models.Post{
    Type: models.Reshare,
    Guid: entity.Guid(),
    PersonID: fetch.Person.ID,
    CreatedAt: createdAt,
    RootPersonID: fetch.Person.ID,
    RootGuid: &rootGuid,
    Public: true,
  }
  err = db.Create(&reshare).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
}

func (receiver *Receiver) StatusMessage(entity federation.MessagePost) {
  var (
    post models.Post
    user models.Person
    person models.Person
  )

  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  err = post.Cast(entity)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  // NOTE ignore if it fails cause
  // if multiple user receive the
  // same private status message you
  // need a shareable item for everyone
  // even if the post already exists
  db.Create(&post)

  if !entity.Public() && receiver.Guid != "" {
    err = db.Where("guid = ?", receiver.Guid).First(&person).Error
    if err != nil {
      revel.AppLog.Error(err.Error(), "guid", receiver.Guid)
      return
    }

    err = db.Find(&user, person.UserID).Error
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }

    shareable := models.Shareable{
      ShareableID: post.ID,
      ShareableType: models.ShareablePost,
      UserID: user.ID,
    }
    err = db.Create(&shareable).Error
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
  }
}
