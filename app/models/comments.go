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
)

type Comment struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  Text string `gorm:"type:text"`
  ShareableID uint `gorm:"size:4"`
  PersonID uint `gorm:"size:4"`
  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Guid string `gorm:"size:191"`
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

  CommentID uint
  AuthorSignature string `gorm:"type:text"`
  SignatureOrderID uint
  AdditionalData string

  SignatureOrder SignatureOrder
}

type CommentSignatures []CommentSignature

// Model Interface Type
//   FetchID() uint
//   FetchGuid() string
//   FetchType() string
//   FetchPersonID() uint
//   FetchText() string
//   HasPublic() bool
//   IsPublic() bool
func (c Comment) FetchID() uint { return c.ID }
func (c Comment) FetchGuid() string { return c.Guid }
func (Comment) FetchType() string { return ShareableComment }
func (c Comment) FetchPersonID() uint { return c.PersonID }
func (c Comment) FetchText() string { return c.Text }
func (Comment) HasPublic() bool { return true }
func (Comment) IsPublic() bool { return false }
// Model Interface Type

func (comment *Comment) AfterFind() error {
  if structLoaded(comment.Person.CreatedAt) {
    return nil
  }

  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Model(comment).Related(&comment.Person).Error
}

func (c *Comment) Count() (count int, err error) {
  db, err := OpenDatabase()
  if err != nil {
    return -1, err
  }
  defer db.Close()

  db.Table("comments").Joins(
    `left join people on comments.person_id = people.id`,
  ).Where("people.user_id > 0").Count(&count)
  return
}

func (c *Comment) AfterSave(db *gorm.DB) error {
  err := searchAndCreateTags(*c, db)
  if err != nil {
    return err
  }

  if _, user, ok := c.ParentPostUser(); ok {
    return user.Notify(*c)
  } else {
    return checkForMentionsInText(*c)
  }
}

func (c *Comment) Create(entity *federation.EntityComment) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  err = c.Cast(entity)
  if err != nil {
    return
  }
  return db.Create(c).Error
}

func (c *Comment) Delete() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  if c.ID == 0 {
    // see http://jinzhu.me/gorm/crud.html#delete
    panic("Setting ID to zero will delete all comment records")
  }
  return db.Delete(c).Error
}

func (c *Comment) AfterDelete(db *gorm.DB) (err error) {
  // aspect_visibilities
  err = db.Where("shareable_id = ? and shareable_type = ?",
    c.ID, ShareableComment).Delete(AspectVisibility{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // shareables
  err = db.Where("shareable_id = ? and shareable_type = ?",
    c.ID, ShareableComment).Delete(Shareable{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // shareable_taggings
  err = db.Where("shareable_id = ? and shareable_type = ?",
    c.ID, ShareableComment).Delete(ShareableTagging{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // notifications
  err = db.Where("shareable_guid = ? and shareable_type = ?",
    c.Guid, ShareableComment).Delete(Notification{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // like_signatures
  return db.Where("like_id = ?", c.ID).Delete(LikeSignature{}).Error
}

func (c *Comment) Cast(entity *federation.EntityComment) (err error) {
  db, err := OpenDatabase()
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
  err = db.Where("author = ?", entity.Author).First(&person).Error
  if err != nil {
    return
  }

  createdAt, err := entity.CreatedAt.Time()
  if err != nil {
    createdAt = time.Now()
  }
  (*c).CreatedAt = createdAt
  (*c).Text = entity.Text
  (*c).ShareableID = post.ID
  (*c).PersonID = person.ID
  (*c).Guid = entity.Guid
  (*c).ShareableType = ShareablePost
  (*c).Signature.AuthorSignature = entity.AuthorSignature
  return nil
}

func (c *Comment) FindByID(id uint) error { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.First(c, id).Error
}

func (c *Comment) FindByGuid(guid string) error { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("guid = ?", guid).First(c).Error
}

func (c *Comment) ParentPostUser() (Post, User, bool) {
  post, ok := c.ParentPost(); if !ok {
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

func (c *Comment) ParentPost() (Post, bool) {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return Post{}, false
  }
  defer db.Close()

  var post Post
  err = db.First(&post, c.ShareableID).Error
  return post, err == nil
}

func (c *Comments) FindByPostID(id uint) (err error) { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("shareable_id = ? and shareable_type = ?", id, ShareablePost).Find(c).Error
}

func (c *Comment) FindByUserAndID(user User, id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  // if no shareable record for this comment
  // exists it is a public message
  if err = c.FindByID(id); err != nil {
    return err
  }
  if db.Where(`shareable_id = ? and shareable_type = ?`,
    c.ShareableID, ShareablePost).First(&Shareable{}).RecordNotFound() {
    return nil
  }

  return db.Joins(`left join shareables
    on shareables.shareable_id = comments.shareable_id`).
    Where(`comments.id = ?
      and shareables.shareable_type = ?
      and shareables.user_id = ?`, id, ShareablePost, user.ID).
    Find(c).Error
}

func (c *Comment) FindByUserAndGuid(user User, guid string) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  // if no shareable record for this comment
  // exists it is a public message
  if err = c.FindByGuid(guid); err != nil {
    return err
  }
  if db.Where(`shareable_id = ? and shareable_type = ?`,
    c.ShareableID, ShareablePost).First(&Shareable{}).RecordNotFound() {
    return nil
  }

  return db.Joins(`left join shareables
    on shareables.shareable_id = comments.shareable_id`).
    Where(`comments.guid = ?
      and shareables.shareable_type = ?
      and shareables.user_id = ?`, guid, ShareablePost, user.ID).
    Find(c).Error
}
