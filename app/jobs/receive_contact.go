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
  "fmt"
  "net/http"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "git.feneas.org/ganggo/ganggo/app/models"
  federation "git.feneas.org/ganggo/federation"
  fhelpers "git.feneas.org/ganggo/federation/helpers"
  api "git.feneas.org/ganggo/api/app"
  "bytes"
)

func (receiver *Receiver) Contact(entity federation.MessageContact) {
  var (
    user models.User
    person models.Person
  )

  username, err := helpers.ParseUsername(entity.Recipient())
  if err != nil {
    revel.AppLog.Error("Receiver Contact", err.Error(), err)
    return
  }

  err = user.FindByUsername(username)
  if err != nil {
    revel.AppLog.Error("Receiver Contact", err.Error(), err)
    return
  }

  err = person.FindByAuthor(entity.Author())
  if err != nil {
    revel.AppLog.Error("Receiver Contact", err.Error(), err)
    return
  }

  contact := models.Contact{
    UserID: user.ID,
    PersonID: person.ID,
    Sharing: entity.Sharing(),
  }

  err = contact.Create()
  if err != nil {
    err = contact.Update()
    if err != nil {
      revel.AppLog.Error("Receiver Contact", err.Error(), err)
      return
    }
  }

  // NOTE ActivityPub needs an accept response
  // XXX how can we integrate this in the federation lib?
  if entity.Type().Proto == federation.ActivityPubProtocol {
    if follow, ok := entity.(*federation.ActivityPubFollow); ok {
      var user models.User
      var person models.Person
      err = user.FindByID(contact.UserID)
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }

      err = person.FindByID(contact.PersonID)
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }

      actor := fmt.Sprintf("%s%s/api/%s/ap/user/%s/actor",
        api.PROTO, api.ADDRESS, api.API_VERSION, user.Username)
      accept := &federation.ActivityPubAccept{
        ActivityPubContext: federation.ActivityPubContext{
          ActivityPubBase: federation.ActivityPubBase{
            Id: actor, Type: federation.ActivityTypeAccept,
          },
        },
        Actor: actor,
        Object: *follow,
      }
      entityXml, err := accept.Marshal(nil, nil)
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }

      header := http.Header{
        "Content-Type": []string{federation.CONTENT_TYPE_JSON},
      }

      privKey, err := fhelpers.ParseRSAPrivateKey(
        []byte(user.SerializedPrivateKey))
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }

      client := (&federation.HttpClient{}).New(actor, privKey)
      err = client.Push(person.Inbox, header, bytes.NewBuffer(entityXml))
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }
    } else {
      revel.AppLog.Error("Cannot cast to EntityFollow!")
      return
    }
  }
}
