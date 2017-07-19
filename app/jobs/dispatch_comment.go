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

const COMMENT_SIG_ORDER = "author guid parent_guid text"

func (d *Dispatcher) Comment(comment *federation.EntityComment) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  defer db.Close()

  guid, err := helpers.Uuid()
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  (*comment).Author = d.User.Person.Author
  (*comment).Guid = guid

  authorSig, err := federation.AuthorSignature(
    *comment, COMMENT_SIG_ORDER, d.User.SerializedPrivateKey,
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
      *comment, COMMENT_SIG_ORDER, parentUser.SerializedPrivateKey,
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

  var visibility models.AspectVisibility
  err = visibility.FindByParentGuid(comment.ParentGuid)
  if err == nil && helpers.IsLocalHandle(d.ParentPerson.Author) {
    sendToAspect(visibility.AspectID, d.User.SerializedPrivateKey, comment.Author, entityXml)
  } else {
    payload, err := federation.MagicEnvelope(
      d.User.SerializedPrivateKey,
      comment.Author, entityXml,
    ); if err != nil {
      revel.ERROR.Println(err)
      return
    }
    sendPublic(payload)
  }
}
