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
  "github.com/ganggo/ganggo/app/models"
  "github.com/ganggo/ganggo/app/helpers"
  federation "github.com/ganggo/federation"
  fhelpers "github.com/ganggo/federation/helpers"
  diaspora "github.com/ganggo/federation/diaspora"
  run "github.com/revel/modules/jobs/app/jobs"
  "bytes"
)

type Dispatcher struct {
  User models.User
  Model interface{}
  Message interface{}
  Relay bool

  Message2 federation.Message
}

type send struct {
  Host, Guid string
  Payload []byte
}

func (dispatcher Dispatcher) Run() {
  //switch entity := dispatcher.Message2.Entity().(type) {
  //case federation.MessageContact:
  //  revel.AppLog.Debug("Starting contact dispatcher")
  //  dispatcher.Contact(entity)
  //  return // XXX
  //}

  switch entity := dispatcher.Message.(type) {
  case diaspora.EntityLike:
    revel.AppLog.Debug("Starting like dispatcher")
    dispatcher.Like(entity)
  case diaspora.EntityComment:
    revel.AppLog.Debug("Starting comment dispatcher")
    dispatcher.Comment(entity)
  case diaspora.EntityRetraction:
    revel.AppLog.Debug("Starting retraction dispatcher")
    dispatcher.Retraction(entity)
  case diaspora.EntityContact:
    revel.AppLog.Debug("Starting contact dispatcher")
    dispatcher.Contact(entity)
  case diaspora.EntityStatusMessage:
    revel.AppLog.Debug("Starting post dispatcher")
    dispatcher.StatusMessage(entity)
  case diaspora.EntityReshare:
    revel.AppLog.Debug("Starting reshare dispatcher")
    dispatcher.Reshare(entity)
  default:
    revel.AppLog.Error("Unknown entity type in dispatcher!")
  }
}

// XXX do not relay the entity to the sender again
func (dispatcher Dispatcher) Send(parentPost models.Post, parentUser models.User, entityXml []byte, orderID uint) {
  if parentPost.ID > 0 && parentUser.ID > 0 {
    revel.AppLog.Debug("Dispatcher we are root host")

    var order models.SignatureOrder
    err := order.FindByID(orderID)
    if err == nil {
      entityXml, err = diaspora.SortByEntityOrder(order.Order, entityXml)
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }
    }

    privKey, err := fhelpers.ParseRSAPrivateKey(
      []byte(parentUser.SerializedPrivateKey))
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }

    // send to public or aspect
    if parentPost.Public {
      payload, err := diaspora.MagicEnvelope(
        privKey, parentUser.Person.Author, entityXml,
      )
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }
      sendPublic(payload)
    } else {
      var visibility models.AspectVisibility
      err := visibility.FindByPost(parentPost)
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }
      sendToAspect(visibility.AspectID, privKey,
        parentUser.Person.Author, entityXml)
    }
  } else if parentPost.ID > 0 {
    revel.AppLog.Debug("Dispatcher send to public or person")

    privKey, err := fhelpers.ParseRSAPrivateKey(
      []byte(dispatcher.User.SerializedPrivateKey))
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }

    // send to public or person
    if parentPost.Public {
      payload, err := diaspora.MagicEnvelope(
        privKey, dispatcher.User.Person.Author, entityXml,
      )
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }
      sendPublic(payload)
    } else {
      pubKey, err := fhelpers.ParseRSAPublicKey(
        []byte(parentPost.Person.SerializedPublicKey))
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }

      payload, err := diaspora.EncryptedMagicEnvelope(
        privKey, pubKey, dispatcher.User.Person.Author, entityXml,
      )
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }

      host, err := helpers.ParseHost(parentPost.Person.Author)
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }

      run.Now(send{
        Host: host,
        Guid: parentPost.Person.Guid,
        Payload: payload,
      })
    }
  }
}

func sendPublic(payload []byte) {
  var pods models.Pods
  err := pods.FindAll()
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  for _, pod := range pods {
    run.Now(send{
      Host: pod.Host,
      Payload: payload,
    })
  }
}

func sendToAspect(aspectID uint, privKey *rsa.PrivateKey, handle string, xml []byte) {
  var aspect models.Aspect
  err := aspect.FindByID(aspectID)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  for _, member := range aspect.Memberships {
    var person models.Person
    err = person.FindByID(member.PersonID)
    if err != nil {
      revel.AppLog.Error(err.Error())
      continue
    }

    pubKey, err := fhelpers.ParseRSAPublicKey(
      []byte(person.SerializedPublicKey))
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }

    payload, err := diaspora.EncryptedMagicEnvelope(
      privKey, pubKey, handle, xml)
    if err != nil {
      revel.AppLog.Error(err.Error())
      continue
    }

    host, err := helpers.ParseHost(person.Author)
    if err != nil {
      revel.AppLog.Error(err.Error())
      continue
    }
    revel.AppLog.Debug("Private request", "guid", person.Guid, "host", host)

    run.Now(send{
      Host: host,
      Guid: person.Guid,
      Payload: payload,
    })
  }
}

func (s send) Run() {
  var err error
  revel.Config.SetSection("ganggo")
  localhost, found := revel.Config.String("address")
  if !found {
    revel.ERROR.Println("No server address configured")
    return
  }

  revel.AppLog.Debug("Sending payload", "guid", s.Guid,
    "host", s.Host, "payload", string(s.Payload))

  // skip own pod
  if s.Host == localhost {
    revel.AppLog.Debug("Skip own pod")
    return
  }

  if s.Guid == "" {
    err = federation.PushToPublic(s.Host, bytes.NewBuffer(s.Payload))
  } else {
    err = federation.PushToPrivate(s.Host, s.Guid, bytes.NewBuffer(s.Payload))
  }
  if err != nil {
    revel.AppLog.Error("Something went wrong while pushing", "host", s.Host, "err", err)
  }
}
