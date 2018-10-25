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
  "time"
  "crypto/rsa"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/models"
  federation "git.feneas.org/ganggo/federation"
  run "github.com/revel/modules/jobs/app/jobs"
  "git.feneas.org/ganggo/federation/helpers"
  "github.com/microcosm-cc/bluemonday"
)

func (dispatcher *Dispatcher) RelayComment(entity federation.MessageComment) {
  var comment models.Comment
  err := comment.FindByGuid(entity.Guid())
  if err != nil {
    revel.AppLog.Error("Dispatcher RelayComment", err.Error(), err)
    return
  }

  parentPost, parentUser, _ := comment.ParentPostUser()
  persons, err := dispatcher.findRecipients(parentPost, parentUser)
  if err != nil {
    revel.AppLog.Error("Dispatcher RelayComment", err.Error(), err)
    return
  }

  priv, err := helpers.ParseRSAPrivateKey(
    []byte(parentUser.SerializedPrivateKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher RelayComment", err.Error(), err)
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
        revel.AppLog.Error("Dispatcher RelayComment", err.Error(), err)
        continue
      }
    } else if person.Pod.Inbox != "" {
      // pod inbox if available
      endpoint = person.Pod.Inbox
    }

    // post cross protocol
    if entity.Type().Proto != person.Pod.Protocol {
      var proto federation.Protocol
      switch entity.Type().Proto {
      case federation.DiasporaProtocol:
        proto = federation.ActivityPubProtocol
      case federation.ActivityPubProtocol:
        proto = federation.DiasporaProtocol
      default:
        revel.AppLog.Warn("Dispatcher RelayComment skipping cross protocol")
        continue
      }

      translate, err := federation.NewMessageComment(proto)
      if err != nil {
        revel.AppLog.Error("Dispatcher RelayComment", err.Error(), err)
        continue
      }
      translate.SetAuthor(parentUser.Person.Author)
      translate.SetGuid(entity.Guid())
      translate.SetParent(parentPost.Guid)
      if !parentPost.Public {
        translate.SetRecipients(recipients)
      }
      createdAt, err := entity.CreatedAt().Time()
      if err != nil {
        createdAt = time.Now()
      }
      translate.SetCreatedAt(createdAt)

      revel.Config.SetSection("ganggo")
      host := revel.Config.StringDefault("proto", "http://") +
        revel.Config.StringDefault("address", "localhost:9000")

      policy := bluemonday.StrictPolicy()
      text := policy.Sanitize(entity.Text())
      text = fmt.Sprintf(
        "%s\n\n---\n\nOriginal post <a href=\"%s/posts/%s#%s\">%s</a>",
        text, host, parentPost.Guid, entity.Guid(), comment.Person.Author)
      translate.SetText(text)

      err = translate.SetSignature(priv)
      if err != nil {
        revel.AppLog.Error("Dispatcher RelayComment", err.Error(), err)
        continue
      }
      // send and retry if it fails the first time
      run.Now(Retry{
        Pod: &person.Pod,
        Send: func() error {
          return translate.Send(endpoint, priv, pub)
        },
      })
    } else {
      // required for a valid envelope signature
      entity.SetAuthor(parentUser.Person.Author)

      // send and retry if it fails the first time
      run.Now(Retry{
        Pod: &person.Pod,
        Send: func() error {
          return entity.Send(endpoint, priv, pub)
        },
      })
    }
  }
}
