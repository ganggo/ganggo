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
  "time"
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

func (d *Dispatcher) Post(post *federation.EntityStatusMessage) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  defer db.Close()

  // create post
  guid, err := helpers.Uuid()
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  (*post).DiasporaHandle = (*d).User.Person.DiasporaHandle
  (*post).Guid = guid
  // set everything to utc
  // otherwise signature fails
  (*post).CreatedAt = time.Now().UTC()
  (*post).ProviderName = "GangGo"
  (*post).Public = true

  // save post locally
  var dbPost models.Post
  err = dbPost.Cast(post, nil)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  err = db.Create(&dbPost).Error
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  // send post to d*
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

  revel.TRACE.Println("USER", d.User)

  payload, err := federation.MagicEnvelope(
    (*d).User.SerializedPrivateKey,
    []byte((*post).DiasporaHandle),
    entityXml,
  )

  // send it to the network
  send(payload)
}
