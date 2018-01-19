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
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  federation "gopkg.in/ganggo/federation.v0"
  "sync"
  "bytes"
)

type Dispatcher struct {
  User models.User
  Model interface{}
  Message interface{}
  Relay bool
}

func (dispatcher *Dispatcher) Run() {
  switch entity := dispatcher.Message.(type) {
  case federation.EntityLike:
    revel.AppLog.Debug("Starting like dispatcher")
    dispatcher.Like(entity)
  case federation.EntityComment:
    revel.AppLog.Debug("Starting comment dispatcher")
    dispatcher.Comment(entity)
  case federation.EntityRetraction:
    revel.AppLog.Debug("Starting retraction dispatcher")
    dispatcher.Retraction(entity)
  case federation.EntityContact:
    revel.AppLog.Debug("Starting contact dispatcher")
    dispatcher.Contact(entity)
  case federation.EntityStatusMessage:
    revel.AppLog.Debug("Starting post dispatcher")
    dispatcher.StatusMessage(entity)
  case federation.EntityReshare:
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
      entityXml = federation.SortByEntityOrder(order.Order, entityXml)
    }

    // send to public or aspect
    if parentPost.Public {
      payload, err := federation.MagicEnvelope(
        parentUser.SerializedPrivateKey,
        parentUser.Person.Author, entityXml,
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
      sendToAspect(visibility.AspectID,
        parentUser.SerializedPrivateKey,
        parentUser.Person.Author, entityXml)
    }
  } else if parentPost.ID > 0 {
    revel.AppLog.Debug("Dispatcher send to public or person")
    // send to public or person
    if parentPost.Public {
      payload, err := federation.MagicEnvelope(
        dispatcher.User.SerializedPrivateKey,
        dispatcher.User.Person.Author, entityXml,
      )
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }
      sendPublic(payload)
    } else {
      payload, err := federation.EncryptedMagicEnvelope(
        dispatcher.User.SerializedPrivateKey,
        parentPost.Person.SerializedPublicKey,
        dispatcher.User.Person.Author, entityXml,
      )
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }

      _, host, err := helpers.ParseAuthor(parentPost.Person.Author)
      if err != nil {
        revel.AppLog.Error(err.Error())
        return
      }
      send(nil, host, parentPost.Person.Guid, payload)
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

  var wg sync.WaitGroup
  for i, pod := range pods {
    wg.Add(1)
    go send(&wg, pod.Host, "", payload)
    // do a maximum of e.g. 20 jobs async
    if i >= MAX_ASYNC_JOBS {
      wg.Wait()
    }
  }
}

func sendToAspect(aspectID uint, priv, handle string, xml []byte) {
  var aspect models.Aspect
  err := aspect.FindByID(aspectID)
  if err != nil {
    revel.ERROR.Println("aspectID", aspectID, err)
    return
  }

  var wg sync.WaitGroup
  for i, member := range aspect.Memberships {
    var person models.Person
    err = person.FindByID(member.PersonID)
    if err != nil {
      revel.ERROR.Println(err)
      continue
    }

    payload, err := federation.EncryptedMagicEnvelope(
      priv, person.SerializedPublicKey, handle, xml,
    ); if err != nil {
      revel.ERROR.Println(err)
      continue
    }

    _, host, err := helpers.ParseAuthor(person.Author)
    if err != nil {
      revel.ERROR.Println(err)
      continue
    }
    revel.TRACE.Println("Private request add", person.Guid, "on", host)

    wg.Add(1)
    go send(&wg, host, person.Guid, payload)
    // do a maximum of e.g. 20 jobs async
    if i >= MAX_ASYNC_JOBS {
      wg.Wait()
    }
  }
}

func send(wg *sync.WaitGroup, host, guid string, payload []byte) {
  var err error
  revel.Config.SetSection("ganggo")
  localhost, found := revel.Config.String("address")
  if !found {
    revel.ERROR.Println("No server address configured")
    return
  }

  revel.AppLog.Debug("Sending payload", "guid", guid,
    "host", host, "payload", string(payload))

  // skip own pod
  if host == localhost {
    revel.AppLog.Debug("Skip own pod")
    return
  }

  if guid == "" {
    err = federation.PushToPublic(host, bytes.NewBuffer(payload))
  } else {
    err = federation.PushToPrivate(host, guid, bytes.NewBuffer(payload))
  }

  if err != nil {
    revel.AppLog.Error(err.Error(), "host", host)
  }
  if wg != nil { wg.Done() }
}
