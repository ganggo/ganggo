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
  "time"
  "github.com/revel/revel"
  federation "gopkg.in/ganggo/federation.v0"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Comment struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  Text string
  ShareableID uint `gorm:"size:4"`
  PersonID uint `gorm:"size:4"`
  Guid string
  LikesCount int `gorm:"size:4"`
  ShareableType string `gorm:"size:60"`

  Signature CommentSignature
  Person Person
}

type Comments []Comment

type CommentSignature struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  CommentId int
  AuthorSignature string
  SignatureOrderID uint
  AdditionalData string

  SignatureOrder SignatureOrder
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

func (c *Comment) ParentIsLocal() (user User, found bool) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  var post Post
  // XXX here we assume every comment is related to post
  // that could be a problem in respect of private messages
  err = db.First(&post, c.ShareableID).Error
  if err != nil {
    return
  }
  db.Model(&post).Related(&post.Person, "Person")

  if post.Person.UserID > 0 {
    err = db.First(&user, post.Person.UserID).Error
    if err != nil {
      return
    }
    db.Model(&user).Related(&user.Person, "Person")
    found = true
    return
  }
  return
}

func (c *Comments) FindByPostID(id uint) (err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("shareable_id = ? and shareable_type = ?", id, ShareablePost).Find(c).Error
}
