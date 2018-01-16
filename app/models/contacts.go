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
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  federation "gopkg.in/ganggo/federation.v0"
)

type Contact struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  UserID uint `gorm:"size:4"`
  PersonID uint `gorm:"size:4"`
  Sharing bool
  Receiving bool
}

type Contacts []Contact

func (c *Contact) Cast(entity *federation.EntityContact) (err error) {
  var recipient User
  var sender Person

  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  username, _, err := helpers.ParseAuthor(entity.Recipient)
  if err != nil {
    return err
  }

  err = db.Where("username = ?", username).First(&recipient).Error
  if err != nil {
    return
  }

  err = db.Where("author = ?", entity.Author).First(&sender).Error
  if err != nil {
    return
  }

  (*c).UserID = recipient.ID
  (*c).PersonID = sender.ID
  (*c).Receiving = entity.Following
  (*c).Sharing = entity.Sharing
  return
}

func (c *Contact) TriggerNotification(guid string) {
  if c.Receiving && c.Sharing {
    notify := Notification{
      ShareableType: ShareableContact,
      ShareableGuid: guid,
      UserID: c.UserID,
      PersonID: c.PersonID,
      Unread: true,
    }
    if err := notify.Create(); err != nil {
      if err := notify.Update(); err != nil {
        revel.AppLog.Error(err.Error())
      }
    }
  }
}
