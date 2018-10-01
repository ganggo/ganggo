package models
//
// GangGo Application Server
// Copyright (C) 2018 Lukas Matt <lukas@zauberstuhl.de>
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

type Visibility struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  ShareableID uint
  PersonID uint
  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  ShareableType string `gorm:"size:191"`

  Person Person
}

type Visibilities []Visibility

func (visibility *Visibility) AfterFind(db *gorm.DB) error {
  err := db.Model(visibility).Related(&visibility.Person).Error
  if err != nil {
    return err
  }
  return nil
}

func (visibility *Visibility) Create() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(visibility).Error
}

func (visibilities *Visibilities) FindByPost(post Post) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("shareable_id = ? and shareable_type = ?",
    post.ID, ShareablePost).Find(visibilities).Error
}
