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
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
  "encoding/xml"
  "sync"
)

func (r *Receiver) RelayComment(user models.User, visibilities models.AspectVisibilities) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  commentEntity := r.Entity.Data.(federation.EntityComment)

  // always generate a parent author signature
  // if the original post is local
  parentSignature, err := federation.AuthorSignature(commentEntity,
    r.Entity.LocalSignatureOrder(), user.SerializedPrivateKey)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  commentEntity.ParentAuthorSignature = parentSignature

  entityCommentXml, err := xml.Marshal(commentEntity)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  for _, visibility := range visibilities {
    var aspect models.Aspect
    err := aspect.FindByID(visibility.AspectID)
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
        user.SerializedPrivateKey,
        person.SerializedPublicKey,
        user.Person.Author,
        entityCommentXml,
      ); if err != nil {
        revel.ERROR.Println(err)
        return
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
}
