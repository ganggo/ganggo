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
  "github.com/ganggo/ganggo/app/models"
  "github.com/ganggo/ganggo/app/helpers"
  "github.com/ganggo/federation"
  "bytes"
  "sync"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
  MAX_ASYNC_JOBS int
)

type Dispatcher struct {
  User models.User
  Message interface{}
}

func (d *Dispatcher) Run() {
  revel.Config.SetSection("ganggo")
  MAX_ASYNC_JOBS = revel.Config.IntDefault("max_async_jobs", 20)

  switch entity := d.Message.(type) {
  case federation.EntityStatusMessage:
    d.Post(&entity)
  case federation.EntityComment:
    d.Comment(&entity)
  default:
    revel.ERROR.Println("Unknown entity type in dispatcher!")
  }
}

func (d *Dispatcher) Comment(comment *federation.EntityComment) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  defer db.Close()

  err = db.First(&(d.User.Person), d.User.PersonID).Error
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  guid, err := helpers.Uuid()
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  (*comment).DiasporaHandle = d.User.Person.DiasporaHandle
  (*comment).Guid = guid

  authorSig, err := federation.AuthorSignature(
    comment,
    (*d).User.SerializedPrivateKey,
  )
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  // parent author signature
  var (
    parentPost models.Post
    parentUser models.User
  )
  db.Where("guid = ?", comment.ParentGuid).First(&parentPost)
  db.First(&parentPost.Person, parentPost.PersonID)
  // if user is local generate a signature
  err = db.First(&parentUser, parentPost.Person.UserID).Error
  if err == nil {
    parentAuthorSig, err := federation.AuthorSignature(
      comment,
      parentUser.SerializedPrivateKey,
    )
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    (*comment).ParentAuthorSignature = parentAuthorSig
  }
  (*comment).AuthorSignature = authorSig

  // save post locally
  var dbComment models.Comment
  err = dbComment.Cast(comment)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  err = db.Create(&dbComment).Error
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  entityXml, err := xml.Marshal(comment)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  revel.TRACE.Println("entityXml", string(entityXml))

  payload, err := federation.MagicEnvelope(
    (*d).User.SerializedPrivateKey,
    []byte((*comment).DiasporaHandle),
    entityXml,
  )

  // send it to the network
  send(payload)
}

func (d *Dispatcher) Post(post *federation.EntityStatusMessage) {
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

  (*post).DiasporaHandle = (*d).User.Person.DiasporaHandle
  (*post).Guid = guid
  // set everything to utc
  // otherwise signature fails
  (*post).CreatedAt = time.Now().UTC()
  (*post).ProviderName = "GangGo"
  (*post).Public = true

  // save post locally
  var dbPost models.Post
  err = dbPost.Cast(post)
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

func send(payload []byte) {
  revel.TRACE.Println("Sending payload", string(payload))

  //pods, err := models.DB.FindPods()
  //if err != nil {
  //  revel.ERROR.Println(err)
  //  return
  //}
  pods := []models.Pod{
    {Host: "192.168.0.173:3000"},
  }

  var wg sync.WaitGroup
  // lets do it \m/ and publish the post
  for i, pod := range pods {
    wg.Add(1)
    // do everything asynchronous
    go func() {
      err := federation.PushXmlToPublic(
        pod.Host, bytes.NewBuffer(payload), true,
      )
      if err != nil {
        revel.ERROR.Println(err, pod.Host)
      }
    }()
    // do a maximum of e.g. 20 jobs async
    if i >= MAX_ASYNC_JOBS {
      wg.Wait()
    }
  }
}
