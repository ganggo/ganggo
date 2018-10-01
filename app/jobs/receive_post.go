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
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "git.feneas.org/ganggo/ganggo/app/models"
  federation "git.feneas.org/ganggo/federation"
)

func (receiver *Receiver) Reshare(entity federation.MessageReshare) {
  fetch := FetchAuthor{Author: entity.ParentAuthor()}
  fetch.Run()
  if fetch.Err != nil {
    revel.AppLog.Error("Receiver Reshare", fetch.Err.Error(), fetch.Err)
    return
  }

  createdAt, err := entity.CreatedAt().Time()
  if err != nil {
    revel.AppLog.Error("Receiver Reshare", err.Error(), err)
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
      revel.AppLog.Error("Receiver Reshare", err.Error(), err)
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
    Protocol: receiver.Message.Type().Proto,
  }
  err = reshare.Create()
  if err != nil {
    revel.AppLog.Error("Receiver Reshare", err.Error(), err)
    return
  }
}

func (receiver *Receiver) StatusMessage(entity federation.MessagePost) {
  var person models.Person
  err := person.FindByAuthor(entity.Author())
  if err != nil {
    revel.AppLog.Error("Receiver StatusMessage", err.Error(), err)
    return
  }

  createdAt, err := entity.CreatedAt().Time()
  if err != nil {
    revel.AppLog.Error("Receiver StatusMessage", err.Error(), err)
    return
  }

  post := models.Post{
    CreatedAt: createdAt,
    Public: entity.Public(),
    Guid: entity.Guid(),
    Type: models.StatusMessage,
    Text: entity.Text(),
    ProviderName: entity.Provider(),
    PersonID: person.ID,
    Protocol: receiver.Message.Type().Proto,
  }
  // NOTE ignore if it fails cause
  // if multiple user receive the
  // same private status message you
  // need a shareable item for everyone
  // even if the post already exists
  post.Create()

  if !entity.Public() && receiver.Guid != "" {
    var localPerson models.Person
    err = localPerson.FindByGuid(receiver.Guid)
    if err != nil {
      revel.AppLog.Error("Receiver StatusMessage", err.Error(), err)
      return
    }

    var user models.Person
    err = user.FindByID(localPerson.UserID)
    if err != nil {
      revel.AppLog.Error("Receiver StatusMessage", err.Error(), err)
      return
    }

    shareable := models.Shareable{
      ShareableID: post.ID,
      ShareableType: models.ShareablePost,
      UserID: user.ID,
    }
    shareable.Create()
    if err != nil {
      revel.AppLog.Error("Receiver StatusMessage", err.Error(), err)
      return
    }

    for _, recipient := range entity.Recipients() {
      if helpers.IsLocalHandle(recipient) {
        // skip local users
        continue
      }

      person, ok := receiver.CheckAuthor(recipient); if !ok {
        // skip persons unkown to the db
        continue
      }

      visibility := models.Visibility{
        ShareableID: post.ID,
        PersonID: person.ID,
        ShareableType: models.ShareablePost,
      }
      err = visibility.Create()
      if err != nil {
        revel.AppLog.Error("Receiver StatusMessage", err.Error(), err)
        continue
      }
    }
  }
}
