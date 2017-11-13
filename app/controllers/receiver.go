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
  "io/ioutil"
)

type Receiver struct {
  *revel.Controller
}

func (r Receiver) Public() revel.Result {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return r.Render()
  }
  defer db.Close()

  request := r.Request.In.GetRaw().(*http.Request)
  content, err := ioutil.ReadAll(request.Body)
  if err != nil {
    r.Response.Status = http.StatusNotAcceptable
    return r.Render()
  }

  revel.TRACE.Println("public", string(content))

  // in case it succeeds reply with status 202
  r.Response.Status = http.StatusAccepted

  message, err := federation.ParseDecryptedRequest(content)
  if err != nil {
    revel.ERROR.Println(err)
    // NOTE Send accept code even tho the entity is not
    // known otherwise the sender pod will throw an error
    //r.Response.Status = http.StatusNotAcceptable
    return r.Render()
  }

  // XXX investigate whether this is a
  // d* problem in production mode as well
  receiverJob := jobs.Receiver{Entity: message.Entity}
  go receiverJob.Run()

  return r.Render()
}

func (r Receiver) Private() revel.Result {
  var (
    guid string
    wrapper federation.AesWrapper
    person models.Person
    user models.User
  )

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return r.Render()
  }
  defer db.Close()

  r.Params.BindJSON(&wrapper)
  r.Params.Bind(&guid, "guid")
  r.Response.Status = http.StatusAccepted

  revel.TRACE.Println("aes request", wrapper)

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

  message, err := federation.ParseEncryptedRequest(wrapper, []byte(user.SerializedPrivateKey))
  if err != nil {
    revel.ERROR.Println(err)
    // NOTE Send accept code even tho the entity is not
    // known otherwise the sender pod will throw an error
    //r.Response.Status = http.StatusNotAcceptable
    return r.Render()
  }

  receiverJob := jobs.Receiver{
    Entity: message.Entity,
    Guid: guid,
  }
  // XXX investigate whether this is a
  // d* problem in production mode as well
  go receiverJob.Run()

  return r.Render()
}
