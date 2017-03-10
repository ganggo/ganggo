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
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

func (d *Dispatcher) Like(like *federation.EntityLike) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  defer db.Close()

  err = db.First(&d.User.Person, d.User.PersonID).Error
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  // create post
  guid, err := helpers.Uuid()
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  (*like).Positive = true
  (*like).Guid = guid
  (*like).DiasporaHandle = (*d).User.Person.DiasporaHandle
  (*like).TargetType = models.ShareablePost

  //(*like).ParentGuid =
  //(*like).AuthorSignature =
  //(*like).ParentAuthorSignature =


  // save post locally

  // send post to d*
  entity := federation.Entity{
    Post: federation.EntityPost{
      Like: like,
    },
  }

  entityXml, err := xml.Marshal(entity)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  revel.TRACE.Println("USER", d.User)

  payload, err := federation.MagicEnvelope(
    (*d).User.SerializedPrivateKey,
    []byte((*like).DiasporaHandle),
    entityXml,
  )

  // send it to the network
  send(payload)
}
