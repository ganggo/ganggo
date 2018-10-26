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
  federation "git.feneas.org/ganggo/federation"
  "git.feneas.org/ganggo/federation/helpers"
  run "github.com/revel/modules/jobs/app/jobs"
)

func (dispatcher *Dispatcher) Comment(comment models.Comment) {
  parentPost, parentUser, _ := comment.ParentPostUser()
  persons, err := dispatcher.findRecipients(parentPost, parentUser)
  if err != nil {
    revel.AppLog.Error("Dispatcher Comment", err.Error(), err)
    return
  }

  priv, err := helpers.ParseRSAPrivateKey(
    []byte(dispatcher.User.SerializedPrivateKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher Comment", err.Error(), err)
    return
  }

  recipients := []string{}
  for _, person := range persons {
    recipients = append(recipients, person.Author)
  }

  for _, person := range persons {
    endpoint := person.Inbox
    var pub *rsa.PublicKey = nil
    if !parentPost.Public {
      pub, err = helpers.ParseRSAPublicKey(
        []byte(person.SerializedPublicKey))
      if err != nil {
        revel.AppLog.Error("Dispatcher Comment", err.Error(), err)
        continue
      }
    } else if person.Pod.Inbox != "" {
      // pod inbox if available
      endpoint = person.Pod.Inbox
    }

    var entity federation.MessageBase
    if dispatcher.Retract {
      retract, err := federation.NewMessageRetract(person.Pod.Protocol)
      if err != nil {
        revel.AppLog.Debug("Dispatcher Comment", err.Error(), err)
        continue
      }
      retract.SetAuthor(comment.Person.Author)
      retract.SetParentType(federation.Comment)
      retract.SetParentGuid(comment.Guid)
      entity = retract
    } else {
      post, err := federation.NewMessageComment(person.Pod.Protocol)
      if err != nil {
        revel.AppLog.Debug("Dispatcher Comment", err.Error(), err)
        continue
      }

      post.SetAuthor(comment.Person.Author)
      if !parentPost.Public {
        post.SetRecipients(recipients)
      }
      post.SetText(comment.Text)
      post.SetGuid(comment.Guid)
      post.SetParent(parentPost.Guid)
      post.SetCreatedAt(comment.CreatedAt)
      err = post.SetSignature(priv)
      if err != nil {
        revel.AppLog.Error("Dispatcher Comment", err.Error(), err)
        continue
      }
      entity = post
    }

    // send and retry if it fails the first time
    run.Now(RetryOnFail{
      Pod: &person.Pod,
      Send: func() error {
        return entity.Send(endpoint, priv, pub)
      },
    })
  }
}
