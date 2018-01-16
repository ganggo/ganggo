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
  ParentUser *models.User
  ParentPerson *models.Person
  Message interface{}
  AspectID uint
}

func (d *Dispatcher) Run() {
  switch entity := d.Message.(type) {
  case federation.EntityStatusMessage:
    d.Post(entity)
  case federation.EntityComment:
    d.Comment(&entity)
  case federation.EntityLike:
    d.Like(entity)
  case federation.EntityContact:
    d.Contact(entity)
  default:
    revel.ERROR.Println("Unknown entity type in dispatcher!")
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

  // skip own pod
  if host == localhost { return }

  revel.TRACE.Println("Sending payload to", guid, host, "with", string(payload))
  if guid == "" {
    err = federation.PushToPublic(host, bytes.NewBuffer(payload))
  } else {
    err = federation.PushToPrivate(host, guid, bytes.NewBuffer(payload))
  }

  if err != nil { revel.ERROR.Println(err, host) }
  if wg != nil { wg.Done() }
}
