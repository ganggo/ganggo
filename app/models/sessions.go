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

type Session struct {
  CreatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Token string `gorm:"size:191"`
  UserID uint `gorm:"size:4"`
  User User
}

type Sessions []Session

func (s *Session) AfterFind(db *gorm.DB) error {
  if structLoaded(s.User.CreatedAt) {
    return nil
  }

  return db.Model(s).Related(&s.User).Error
}

func (s *Sessions) FindByTimeRange(from, to time.Time) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("created_at between ? and ?", from, to).Find(s).Error
}

func (s *Sessions) Delete() error {
  for _, session := range *s {
    err := session.Delete()
    if err != nil {
      return err
    }
  }
  return nil
}

func (s *Session) Delete() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  if s.Token == "" {
    panic("Cannot delete empty session struct!")
  }

  return db.Where("token = ?", s.Token).Delete(s).Error
}
