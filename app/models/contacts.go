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
  "github.com/ganggo/ganggo/app/helpers"
  "github.com/ganggo/federation"
  "github.com/jinzhu/gorm"
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

func (c *Contact) Cast(entity *federation.EntityRequest) (err error) {
  var (
    recipient User
    sender Person
  )

  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return err
  }
  defer db.Close()

  username, _, err := helpers.ParseDiasporaHandle(entity.Recipient)
  if err != nil {
    return err
  }


  err = db.Where("username = ?", username).First(&recipient).Error
  if err != nil {
    return
  }

  err = db.Where("diaspora_handle = ?", entity.Sender).First(&sender).Error
  if err != nil {
    return
  }

  (*c).UserID = recipient.ID
  (*c).PersonID = sender.ID
  (*c).Receiving = true

  return
}
