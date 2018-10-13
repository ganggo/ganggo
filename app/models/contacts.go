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
  "git.feneas.org/ganggo/gorm"
)

type Contact struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  UserID uint `gorm:"size:4"`
  PersonID uint `gorm:"size:4"`
  Sharing bool
}

type Contacts []Contact

// Model Interface Type
//   FetchID() uint
//   FetchGuid() string
//   FetchType() string
//   FetchPersonID() uint
//   FetchText() string
//   HasPublic() bool
//   IsPublic() bool
func (c Contact) FetchID() uint { return c.ID }
func (c Contact) FetchGuid() string {
  var person Person
  err := person.FindByID(c.PersonID)
  if err != nil {
    panic(err.Error())
  }
  return person.Guid
}
func (Contact) FetchType() string { return ShareableContact }
func (c Contact) FetchPersonID() uint { return c.PersonID }
func (Contact) FetchText() string { return "" }
func (Contact) HasPublic() bool { return false }
func (Contact) IsPublic() bool { return false }
// Model Interface Type

func (c *Contact) AfterSave(db *gorm.DB) error {
  if c.Sharing {
    var user User
    err := user.FindByID(c.UserID)
    if err != nil {
      return err
    }
    return user.Notify(*c)
  }
  return nil
}

func (c *Contact) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(c).Error
}

func (c *Contact) Update() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  sharing := c.Sharing
  if c.ID <= 0 {
    err = db.Where("user_id = ? AND person_id = ?",
      c.UserID, c.PersonID).First(c).Error
    if err != nil {
      return err
    }
  }

  return db.Model(c).Update("sharing", sharing).Error
}

func (c *Contacts) FindByUserID(id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ?", id).Find(c).Error
}
