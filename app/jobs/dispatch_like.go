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

func (dispatcher *Dispatcher) Like(like models.Like) {
  parentPost, parentUser, _ := like.ParentPostUser()
  persons, err := dispatcher.findRecipients(parentPost, parentUser)
  if err != nil {
    revel.AppLog.Error("Dispatcher Like", err.Error(), err)
    return
  }

  priv, err := helpers.ParseRSAPrivateKey(
    []byte(dispatcher.User.SerializedPrivateKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher Like", err.Error(), err)
    return
  }

  for _, person := range persons {
    var pub *rsa.PublicKey = nil
    endpoint := person.Inbox

    if !parentPost.Public {
      pub, err = helpers.ParseRSAPublicKey(
        []byte(parentPost.Person.SerializedPublicKey))
      if err != nil {
        revel.AppLog.Error("Dispatcher Like", err.Error(), err)
        return
      }
    } else if person.Pod.Inbox != "" {
      // pod inbox if available
      endpoint = person.Pod.Inbox
    }

    var entity federation.MessageBase
    if dispatcher.Retract {
      retract, err := federation.NewMessageRetract(person.Pod.Protocol)
      if err != nil {
        revel.AppLog.Error("Dispatcher Like", err.Error(), err)
        continue
      }
      retract.SetAuthor(dispatcher.User.Person.Author)
      retract.SetParentType(federation.Like)
      retract.SetParentGuid(like.Guid)
      entity = retract
    } else {
      post, err := federation.NewMessageLike(person.Pod.Protocol)
      if err != nil {
        revel.AppLog.Error("Dispatcher Like", err.Error(), err)
        return
      }
      post.SetAuthor(dispatcher.User.Person.Author)
      post.SetGuid(like.Guid)
      post.SetPositive(like.Positive)
      post.SetParent(parentPost.Guid)
      err = post.SetSignature(priv)
      if err != nil {
        revel.AppLog.Error("Dispatcher Like", err.Error(), err)
        continue
      }
      entity = post
    }

    // send and retry if it fails the first time
    run.Now(Retry{
      Pod: &person.Pod,
      Send: func() error {
        return entity.Send(endpoint, priv, pub)
      },
    })
  }
}
