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

import "github.com/jinzhu/gorm"

type User struct {
  gorm.Model

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Username string `gorm:"size:191"`
  Email string `gorm:"size:191"`
  SerializedPrivateKey string `gorm:"type:text"`
  EncryptedPassword string

  PersonID uint
  Person Person `gorm:"ForeignKey:PersonID"`

  Aspects []Aspect `gorm:"AssociationForeignKey:UserID"`
}

func (user *User) AfterCreate(tx *gorm.DB) error {
  return tx.Model(&user.Person).Update("user_id", user.ID).Error
}

func (user *User) AfterFind(db *gorm.DB) error {
  if structLoaded(user.Person.CreatedAt) {
    return nil
  }

  err := db.Model(user).Related(&user.Person).Error
  if err != nil {
    return err
  }

  return db.Model(user).Related(&user.Aspects).Error
}

func (user *User) FindByID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(user, id).Error
}

func (user *User) FindByUsername(name string) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("username = ?", name).Find(user).Error
}

func (user *User) Count() (count int, err error) {
  db, err := OpenDatabase()
  if err != nil {
    return -1, err
  }
  defer db.Close()

  db.Table("users").Count(&count)
  return
}

func (user *User) Notify(model Model) error {
  // do not send notification for your own activity
  if user.Person.ID == model.FetchPersonID() {
    return nil
  }

  notify := Notification{
    ShareableType: model.FetchType(),
    ShareableGuid: model.FetchGuid(),
    UserID: user.ID,
    PersonID: model.FetchPersonID(),
    Unread: true,
  }
  return notify.Create()
}
