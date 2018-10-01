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
  "crypto/rsa"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/federation"
  fhelpers "git.feneas.org/ganggo/federation/helpers"
)

func (dispatcher *Dispatcher) StatusMessage(post models.Post) {
  persons, err := dispatcher.findRecipients(&post, &dispatcher.User)
  if err != nil {
    revel.AppLog.Error("Dispatcher StatusMessage", err.Error(), err)
    return
  }

  priv, err := fhelpers.ParseRSAPrivateKey(
    []byte(dispatcher.User.SerializedPrivateKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher StatusMessage", err.Error(), err)
    return
  }

  recipients := []string{}
  for _, person := range persons {
    recipients = append(recipients, person.Author)
  }

  for _, person := range persons {
    var pub *rsa.PublicKey = nil
    endpoint := person.Inbox
    if !post.Public {
      pub, err = fhelpers.ParseRSAPublicKey(
        []byte(person.SerializedPublicKey))
      if err != nil {
        revel.AppLog.Error("Dispatcher StatusMessage", err.Error(), err)
        continue
      }
    } else if person.Pod.Inbox != "" {
      // pod inbox if available
      endpoint = person.Pod.Inbox
    }

    var entity federation.MessageBase
    if dispatcher.Retract {
      msg, err := federation.NewMessageRetract(person.Pod.Protocol)
      if err != nil {
        revel.AppLog.Error("Dispatcher StatusMessage", err.Error(), err)
        return
      }
      msg.SetAuthor(post.Person.Author)
      msg.SetParentType(federation.StatusMessage)
      msg.SetParentGuid(post.Guid)
      entity = msg
    } else if post.Type == models.StatusMessage {
      msg, err := federation.NewMessagePost(person.Pod.Protocol)
      if err != nil {
        revel.AppLog.Error("Dispatcher StatusMessage", err.Error(), err)
        continue
      }
      msg.SetAuthor(post.Person.Author)
      if !post.Public {
        msg.SetRecipients(recipients)
      }
      msg.SetText(post.Text)
      msg.SetGuid(post.Guid)
      msg.SetPublic(post.Public)
      msg.SetCreatedAt(post.CreatedAt)
      entity = msg
    } else if post.Type == models.Reshare {
      var parent models.Person
      err = parent.FindByID(post.RootPersonID)
      if err != nil {
        revel.AppLog.Error("Dispatcher Reshare", err.Error(), err)
        continue
      }

      msg, err := federation.NewMessageReshare(person.Pod.Protocol)
      if err != nil {
        revel.AppLog.Error("Dispatcher StatusMessage", err.Error(), err)
        continue
      }
      msg.SetAuthor(post.Person.Author)
      msg.SetGuid(post.Guid)
      msg.SetParentAuthor(parent.Author)
      msg.SetParent(*post.RootGuid)
      msg.SetCreatedAt(post.CreatedAt)
      entity = msg
    } else {
      panic("Something went wrong!")
    }

    err = entity.Send(endpoint, priv, pub)
    if err != nil {
      revel.AppLog.Error("Dispatcher Reshare", err.Error(), err)
      continue
    }
  }
}
