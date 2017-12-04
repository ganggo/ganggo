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

import "time"

type Notification struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  TargetType string `gorm:"size:191"`
  TargetGuid string `gorm:"size:191"`
  UserID uint
  User User `json:"-"`
  PersonID uint
  Person Person `json:"-"`
  Unread bool
}

type Notifications []Notification

func (n *Notification) AfterFind() error {
  if structLoaded(n.User.CreatedAt) {
    return nil
  }

  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.Model(n).Related(&n.User).Error
  if err != nil {
    return err
  }
  return db.Model(n).Related(&n.Person).Error
}

func (n *Notifications) FindUnreadByUserID(id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ? and unread = ?", id, true).Find(n).Error
}

func (n *Notification) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(n).Error
}
