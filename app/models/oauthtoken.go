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
  "gopkg.in/ganggo/gorm.v2"
)

type OAuthToken struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  ClientID string `gorm:"size:191"`
  Token string `gorm:"size:191"`

  UserID uint
  User User
}

type OAuthTokens []OAuthToken

func (o *OAuthToken) AfterFind(db *gorm.DB) error {
  if structLoaded(o.User.CreatedAt) {
    return nil
  }

  return db.Model(o).Related(&o.User).Error
}

func (o *OAuthToken) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(o).Error
}

func (o *OAuthToken) Delete(user User) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ?", user.ID).Delete(o).Error
}

func (o *OAuthToken) FindByUserIDAndClientID(userID uint, clientID string) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ? and client_id = ?", userID, clientID).First(o).Error
}

func (o *OAuthToken) FindByToken(token string) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("token = ?", token).First(o).Error
}

func (o *OAuthTokens) FindByUserID(id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ?", id).Find(o).Error
}
