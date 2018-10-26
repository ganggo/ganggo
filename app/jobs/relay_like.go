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

func (dispatcher *Dispatcher) RelayLike(entity federation.MessageLike) {
  var like models.Like
  err := like.FindByGuid(entity.Guid())
  if err != nil {
    revel.AppLog.Error("Dispatcher RelayLike", err.Error(), err)
    return
  }

  parentPost, parentUser, _ := like.ParentPostUser()
  persons, err := dispatcher.findRecipients(parentPost, parentUser)
  if err != nil {
    revel.AppLog.Error("Dispatcher RelayLike", err.Error(), err)
    return
  }

  priv, err := helpers.ParseRSAPrivateKey(
    []byte(parentUser.SerializedPrivateKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher RelayLike", err.Error(), err)
    return
  }

  for _, person := range persons {
    if entity.Type().Proto != person.Pod.Protocol {
      revel.AppLog.Debug("Dispatcher RelayLike skipping cross protocol")
      continue
    }

    endpoint := person.Inbox
    var pub *rsa.PublicKey = nil
    if !parentPost.Public {
      pub, err = helpers.ParseRSAPublicKey(
        []byte(person.SerializedPublicKey))
      if err != nil {
        revel.AppLog.Error("Dispatcher RelayLike", err.Error(), err)
        continue
      }
    } else if person.Pod.Inbox != "" {
      // pod inbox if available
      endpoint = person.Pod.Inbox
    }

    // required for a valid envelope signature
    entity.SetAuthor(parentUser.Person.Author)

    // send and retry if it fails the first time
    run.Now(RetryOnFail{
      Pod: &person.Pod,
      Send: func() error {
        return entity.Send(endpoint, priv, pub)
      },
    })
  }
}
