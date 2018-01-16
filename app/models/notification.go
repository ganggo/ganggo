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
  "github.com/jinzhu/gorm"
)

type Notification struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  ShareableType string `gorm:"size:191"`
  ShareableGuid string `gorm:"size:191"`
  UserID uint
  PersonID uint
  Unread bool

  Person Person `json:"-"`
  User User `json:"-"`
  Comment Comment `json:"-"`
  Post Post `json:"-"`
}

type Notifications []Notification

func (n *Notification) AfterFind(db *gorm.DB) error {
  err := db.Model(n).Related(&n.User).Error
  if err != nil {
    return err
  }

  err = db.Model(n).Related(&n.Person).Error
  if err != nil {
    return err
  }

  if n.ShareableType == ShareablePost {
    var post Post
    err = post.FindByGuid(n.ShareableGuid)
    if err != nil {
      return err
    }
    (*n).Post = post
  } else if n.ShareableType == ShareableComment {
    var comment Comment
    err = comment.FindByGuid(n.ShareableGuid)
    if err != nil {
      return err
    }
    (*n).Comment = comment
  }
  return nil
}

func (n *Notifications) FindUnreadByUserID(id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ? and unread = ?", id, true).Find(n).Error
}

func (n *Notifications) FindByUserID(id uint, offset int) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Offset(offset).Limit(10).
    Where("user_id = ?", id).Find(n).Error
}

func (n *Notification) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(n).Error
}

func (n *Notification) Update() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  if n.ID > 0 {
    return db.Model(n).Update(
      "unread", n.Unread,
    ).Error
  }

  return db.Model(n).Where("user_id = ? and person_id = ?",
    n.UserID, n.PersonID,
  ).Update("unread", n.Unread).Error
}
