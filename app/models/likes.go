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
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Like struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  Positive bool
  TargetID uint `gorm:"size:4"`
  PersonID uint `gorm:"size:4"`
  Guid string
  AuthorSignature string
  TargetType string `gorm:"size:60"`
}

type Likes []Like

func (l *Like) Create(entity *federation.EntityLike) (err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
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

func (l *Like) Cast(entity *federation.EntityLike) (err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
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

  err = db.Where("diaspora_handle = ?",
    entity.DiasporaHandle).First(&person).Error
  if err != nil {
    return
  }

  (*l).Positive = entity.Positive
  (*l).TargetID = post.ID
  (*l).PersonID = person.ID
  (*l).Guid = entity.Guid
  (*l).AuthorSignature = entity.AuthorSignature
  (*l).TargetType = entity.TargetType

  return
}

func (l *Likes) FindByPostID(id uint) (err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("target_id = ? and target_type = ?", id, ShareablePost).Find(l).Error
}
