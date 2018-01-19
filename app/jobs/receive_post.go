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

func (receiver *Receiver) Reshare(entity federation.EntityReshare) {
  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  fetch := FetchAuthor{Author: entity.Author}
  fetch.Run()
  if fetch.Err != nil {
    revel.AppLog.Error(fetch.Err.Error())
    return
  }

  fetchRoot := FetchAuthor{Author: entity.RootAuthor}
  fetchRoot.Run()
  if fetchRoot.Err != nil {
    revel.AppLog.Error(fetchRoot.Err.Error())
    return
  }

  var post models.Post
  if err = post.FindByGuid(entity.RootGuid); err == nil {
    reshare := models.Post{
      Type: models.Reshare,
      Guid: entity.Guid,
      PersonID: fetch.Person.ID,
      CreatedAt: entity.CreatedAt.Time,
      RootPersonID: fetchRoot.Person.ID,
      RootGuid: &entity.RootGuid,
      Public: true,
    }
    err = db.Create(&reshare).Error
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
  }
}

func (receiver *Receiver) StatusMessage(entity federation.EntityStatusMessage) {
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

  fetch := FetchAuthor{Author: entity.Author}
  fetch.Run()
  if fetch.Err != nil {
    revel.AppLog.Error(fetch.Err.Error())
    return
  }

  err = post.Cast(&entity)
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

  if receiver.Guid != "" {
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
