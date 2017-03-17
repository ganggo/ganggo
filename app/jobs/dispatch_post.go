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
  "encoding/xml"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  federation "gopkg.in/ganggo/federation.v0"
  "sync"
)

func (d *Dispatcher) Post(post *federation.EntityStatusMessage) {
  entity := federation.Entity{
    Post: federation.EntityPost{
      StatusMessage: post,
    },
  }

  entityXml, err := xml.Marshal(entity)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  if d.AspectID > 0 {
    var aspect models.Aspect
    err = aspect.FindByID(d.AspectID)
    if err != nil {
      revel.ERROR.Println(err)
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
        (*d).User.SerializedPrivateKey,
        person.SerializedPublicKey,
        (*post).DiasporaHandle,
        entityXml,
      )
      if err != nil {
        revel.ERROR.Println(err)
        continue
      }

      _, host, err := helpers.ParseDiasporaHandle(person.DiasporaHandle)
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
  } else {
    payload, err := federation.MagicEnvelope(
      (*d).User.SerializedPrivateKey,
      (*post).DiasporaHandle,
      entityXml,
    )
    if err != nil {
      revel.ERROR.Println(err)
      return
    }

    sendPublic(payload)
  }
}
