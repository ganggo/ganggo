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
  run "github.com/revel/modules/jobs/app/jobs"
  "git.feneas.org/ganggo/federation/helpers"
)

func (dispatcher *Dispatcher) RelayRetraction(entity federation.MessageRetract) {
  var persons []models.Person
  var serializedPrivateKey string
  var public bool

  switch entity.ParentType() {
  case federation.Reshare:
    var post models.Post
    err := post.FindByGuid(entity.ParentGuid())
    if err != nil {
      revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
      return
    }
    if user, ok := post.IsLocal(); ok {
      public = post.Public
      serializedPrivateKey = user.SerializedPrivateKey
      entity.SetAuthor(user.Person.Author)
      persons, err = dispatcher.findRecipients(&post, &user)
      if err != nil {
        revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
        return
      }
    }
  case federation.StatusMessage:
    var post models.Post
    err := post.FindByGuid(entity.ParentGuid())
    if err != nil {
      revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
      return
    }
    if user, ok := post.IsLocal(); ok {
      public = post.Public
      serializedPrivateKey = user.SerializedPrivateKey
      entity.SetAuthor(user.Person.Author)
      persons, err = dispatcher.findRecipients(&post, &user)
      if err != nil {
        revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
        return
      }
    }
  case federation.Comment:
    var comment models.Comment
    err := comment.FindByGuid(entity.ParentGuid())
    if err != nil {
      revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
      return
    }
    post, user, _ := comment.ParentPostUser()
    public = post.Public
    serializedPrivateKey = user.SerializedPrivateKey
    entity.SetAuthor(user.Person.Author)
    persons, err = dispatcher.findRecipients(post, user)
    if err != nil {
      revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
      return
    }
  case federation.Like:
    var like models.Like
    err := like.FindByGuid(entity.ParentGuid())
    if err != nil {
      revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
      return
    }
    post, user, _ := like.ParentPostUser()
    public = post.Public
    serializedPrivateKey = user.SerializedPrivateKey
    entity.SetAuthor(user.Person.Author)
    persons, err = dispatcher.findRecipients(post, user)
    if err != nil {
      revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
      return
    }
  default:
    revel.AppLog.Error("Unkown TargetType in Dispatcher",
      "retraction", entity)
    return
  }

  priv, err := helpers.ParseRSAPrivateKey([]byte(serializedPrivateKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
    return
  }

  for _, person := range persons {
    if entity.Type().Proto != person.Pod.Protocol {
      // not supported
      revel.AppLog.Debug("Dispatcher RelayRetraction skipping cross protocol")
      continue
    }

    endpoint := person.Inbox
    var pub *rsa.PublicKey = nil
    if !public {
      pub, err = helpers.ParseRSAPublicKey(
        []byte(person.SerializedPublicKey))
      if err != nil {
        revel.AppLog.Error("Dispatcher RelayRetraction", err.Error(), err)
        continue
      }
    } else if person.Pod.Inbox != "" {
      // pod inbox if available
      endpoint = person.Pod.Inbox
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
