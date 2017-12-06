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
  federation "gopkg.in/ganggo/federation.v0"
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

  CommentID int
  AuthorSignature string `gorm:"type:text"`
  SignatureOrderID uint
  AdditionalData string

  SignatureOrder SignatureOrder
}

type CommentSignatures []CommentSignature

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

func (c *Comment) AfterCreate() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  // batch insert doesn't work for gorm, yet
  // see https://github.com/jinzhu/gorm/issues/255
  tags, err := generateTags(c)
  if err == nil && len(tags) > 0 {
    for _, tag := range tags {
      var cnt int
      db.Where("name = ?", tag.Name).Find(&tag).Count(&cnt)
      // if tag already exists skip it
      // and create taggings only
      if cnt == 0 {
        err = db.Create(&tag).Error
        if err != nil {
          return err
        }
      } else {
        for _, shareable := range tag.ShareableTaggings {
          shareable.TagID = tag.ID
          err = db.Create(&shareable).Error
          if err != nil {
            return err
          }
        }
      }
    }
  } else if err != nil {
    return err
  }

  notify, err := generateNotifications(c)
  if err == nil && len(notify) > 0 {
    for _, n := range notify {
      err = db.Create(&n).Error
      if err != nil {
        return err
      }
    }
  }
  return err
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

func (c *Comment) ParentIsLocal() (User, bool) {
  return parentIsLocal(c.ShareableID)
}

func (c *Comments) FindByPostID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("shareable_id = ? and shareable_type = ?", id, ShareablePost).Find(c).Error
}
