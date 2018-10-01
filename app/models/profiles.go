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
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/gorm"
)

type Profile struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  PersonID uint `gorm:"size:4"`
  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Author string `gorm:"size:191"`
  ImageUrl string
  Public bool

  FirstName string `gorm:"null"`
  LastName string `gorm:"null"`
  Birthday time.Time `gorm:"null"`
  Gender string `gorm:"null"`
  Bio string `gorm:"type:text;null"`
  Searchable bool
  Location string `gorm:"null"`
  FullName string `gorm:"size:70"`
  Nsfw bool
}

func (p Profile) Nickname() (nickname string) {
  nickname, _ = helpers.ParseUsername(p.Author)
  return
}

func (profile *Profile) CreateOrUpdate() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  var dbProfile Profile
  err = dbProfile.FindByPersonID(profile.PersonID)
  if err == gorm.ErrRecordNotFound {
    return db.Create(profile).Error
  }
  profile.ID = dbProfile.ID
  return db.Save(profile).Error
}

func (profile *Profile) FindByPersonID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("person_id = ?", id).First(profile).Error
}

func (profile *Profile) FindByAuthor(author string) error {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return err
  }
  defer db.Close()

  return db.Where("author = ?", author).First(profile).Error
}
