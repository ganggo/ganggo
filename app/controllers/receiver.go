package controllers
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
  "net/http"
  "github.com/revel/revel"
  federation "gopkg.in/ganggo/federation.v0"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/jobs"
)

type Receiver struct {
  *revel.Controller
}

func (r Receiver) Public() revel.Result {
  var xml string

  revel.TRACE.Println("REQUEST", r.Request, r.Request.ContentType)

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return r.Render()
  }
  defer db.Close()

  r.Params.Bind(&xml, "xml")
  r.Response.ContentType = "application/json"

  request, err := federation.PreparePublicRequest(xml)
  if err != nil {
    r.Response.Status = http.StatusNotAcceptable
    return r.Render()
  }

  // NOTE investigate whether this is a
  // d* problem in production mode as well
  go func() {
    // check if author is already known
    fetchAuthor := jobs.FetchAuthor{
      Author: request.Header.AuthorId,
    }
    fetchAuthor.Run()
    if fetchAuthor.Err != nil {
      revel.ERROR.Println(err)
      return
    }

    entity, err := request.Parse(
      []byte(fetchAuthor.Person.SerializedPublicKey))
    if err != nil {
      revel.ERROR.Println(err)
      return
    }

    receiverJob := jobs.Receiver{
      Entity: entity,
    }
    go receiverJob.Run()
  }()
  return r.Render()
}

func (r Receiver) Private() revel.Result {
  var (
    guid string
    xml string
    person models.Person
    user models.User
  )

  revel.TRACE.Println("REQUEST", r.Request, r.Request.ContentType)

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return r.Render()
  }
  defer db.Close()

  r.Params.Bind(&xml, "xml")
  r.Params.Bind(&guid, "guid")
  r.Response.ContentType = "application/json"

  err = db.Where("guid like ?", guid).First(&person).Error
  if err != nil {
    revel.ERROR.Println(err)
    // diaspora will not process on StatusNotAcceptable
    return r.Render()
  }

  err = db.First(&user, person.UserID).Error
  if err != nil {
    revel.ERROR.Println(err)
    r.Response.Status = http.StatusNotAcceptable
    return r.Render()
  }

  request, err := federation.PreparePrivateRequest(xml,
    []byte(user.SerializedPrivateKey))
  if err != nil {
    revel.ERROR.Println(err)
    r.Response.Status = http.StatusNotAcceptable
    return r.Render()
  }

  // XXX investigate whether this is a
  // d* problem in production mode as well
  go func() {
    // check if author is already known
    fetchAuthor := jobs.FetchAuthor{
      Author: request.DecryptedHeader.AuthorId,
    }
    fetchAuthor.Run()
    if fetchAuthor.Err != nil {
      revel.ERROR.Println(err)
      return
    }

    entity, err := request.ParsePrivate(
      []byte(fetchAuthor.Person.SerializedPublicKey))
    if err != nil {
      revel.ERROR.Println(err)
      return
    }

    receiverJob := jobs.Receiver{
      Entity: entity,
      Guid: guid,
    }
    go receiverJob.Run()
  }()
  return r.Render()
}
