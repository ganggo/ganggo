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
  "gopkg.in/ganggo/gorm.v2"
)

type Like struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  Positive bool
  ShareableID uint `gorm:"size:4"`
  PersonID uint `gorm:"size:4"`
  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Guid string `gorm:"size:191"`
  ShareableType string `gorm:"size:60"`

  Signature LikeSignature
}

type Likes []Like

type LikeSignature struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  LikeID uint
  AuthorSignature string `gorm:"type:text"`
  SignatureOrderID uint
  AdditionalData string

  SignatureOrder SignatureOrder
}

type LikeSignatures []LikeSignature

// Model Interface Type
//   FetchID() uint
//   FetchGuid() string
//   FetchType() string
//   FetchPersonID() uint
//   FetchText() string
//   HasPublic() bool
//   IsPublic() bool
func (l Like) FetchID() uint { return l.ID }
func (l Like) FetchGuid() string { return l.Guid }
func (Like) FetchType() string { return ShareableLike }
func (l Like) FetchPersonID() uint { return l.PersonID }
func (Like) FetchText() string { return "" }
func (Like) HasPublic() bool { return false }
func (Like) IsPublic() bool { return false }
// Model Interface Type

func (l *Like) Create(entity *federation.EntityLike) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  err = l.Cast(entity)
  if err != nil {
    return
  }
  return db.Create(l).Error
}

func (l *Like) AfterSave(db *gorm.DB) (err error) {
  if _, user, ok := l.ParentPostUser(); ok {
    return user.Notify(*l)
  }
  return nil
}

func (l *Like) Delete() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  if l.ID == 0 {
    // see http://jinzhu.me/gorm/crud.html#delete
    panic("Setting ID to zero will delete all comment records")
  }
  return db.Delete(l).Error
}

func (l *Like) AfterDelete(db *gorm.DB) (err error) {
  // aspect_visibilities
  err = db.Where("shareable_id = ? and shareable_type = ?",
    l.ID, ShareableLike).Delete(AspectVisibility{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // shareables
  err = db.Where("shareable_id = ? and shareable_type = ?",
    l.ID, ShareableLike).Delete(Shareable{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // shareable_taggings
  err = db.Where("shareable_id = ? and shareable_type = ?",
    l.ID, ShareableLike).Delete(ShareableTagging{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // notifications
  err = db.Where("shareable_guid = ? and shareable_type = ?",
    l.Guid, ShareableLike).Delete(Notification{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // like_signatures
  return db.Where("like_id = ?", l.ID).Delete(LikeSignature{}).Error
}

func (l *Like) Cast(entity *federation.EntityLike) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  var (
    post Post
    person Person
  )

  err = db.Where("guid = ?", entity.ParentGuid).First(&post).Error
  if err != nil {
    return
  }

  err = db.Where("author = ?", entity.Author).First(&person).Error
  if err != nil {
    return
  }

  (*l).Positive = entity.Positive
  (*l).ShareableID = post.ID
  (*l).PersonID = person.ID
  (*l).Guid = entity.Guid
  (*l).Signature.AuthorSignature = entity.AuthorSignature
  (*l).ShareableType = entity.ParentType

  return
}

func (l *Like) ParentPostUser() (Post, User, bool) {
  post, ok := l.ParentPost(); if !ok {
    return post, User{}, ok
  }
  if post.Person.UserID <= 0 {
    return post, User{}, false
  }

  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return post, User{}, false
  }
  defer db.Close()

  var user User
  err = user.FindByID(post.Person.UserID)
  return post, user, err == nil
}

func (l *Like) ParentPost() (Post, bool) {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return Post{}, false
  }
  defer db.Close()

  var post Post
  err = db.First(&post, l.ShareableID).Error
  return post, err == nil
}

func (l *Like) FindByUserAndPostID(user User, id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where(`shareable_id = ?
    and shareable_type = ?
    and person_id = ?`,
    id, ShareablePost, user.Person.ID,
  ).First(l).Error
}

func (l *Likes) FindByPostID(id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("shareable_id = ? and shareable_type = ?", id, ShareablePost).Find(l).Error
}

func (l *Like) FindByID(id uint) error { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(l, id).Error
}

func (l *Like) FindByGuid(guid string) error { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("guid = ?", guid).Find(l).Error
}
