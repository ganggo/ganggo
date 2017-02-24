package models
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
  "github.com/ganggo/federation"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Comment struct {
  gorm.Model

  Text string
  ShareableID uint
  PersonID uint
  Guid string
  LikesCount int
  ShareableType string
  Signature CommentSignature
}

type Comments []Comment

type CommentSignature struct {
  gorm.Model

  CommentId int `gorm:"primary_key"`
  AuthorSignature string
  // TODO
  //SignatureOrderId int `gorm:"primary_key"`
  AdditionalData string
}

type CommentSignatures []CommentSignature

func (c *Comment) Cast(entity *federation.EntityComment) (err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return
  }
  defer db.Close()

  var post Post
  err = db.Where("guid = ?", entity.ParentGuid).First(&post).Error
  if err != nil {
    return
  }
  var person Person
  err = db.Where("diaspora_handle = ?",
    entity.DiasporaHandle).First(&person).Error
  if err != nil {
    return
  }

  (*c).Text = entity.Text
  (*c).ShareableID = post.ID
  (*c).PersonID = person.ID
  (*c).Guid = entity.Guid
  (*c).ShareableType = ShareablePost
  (*c).Signature = CommentSignature{
    AuthorSignature: entity.AuthorSignature,
  }
  return nil
}
