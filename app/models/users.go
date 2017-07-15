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
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type User struct {
  gorm.Model

  Username string
  Email string
  SerializedPrivateKey string `gorm:"type:text"`
  EncryptedPassword string

  PersonID uint
  Person Person

  Aspects []Aspect `gorm:"AssociationForeignKey:UserID"`
}

func (user *User) FindByID(id uint) (err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(user, id).Error
}

func (user *User) Count() (count int, err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return -1, err
  }
  defer db.Close()

  db.Table("users").Count(&count)
  return
}
