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
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  federation "gopkg.in/ganggo/federation.v0"
)

type Profile struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Author string `gorm:"size:191"`
  FirstName string `gorm:"null"`
  LastName string `gorm:"null"`
  ImageUrl string
  ImageUrlSmall string
  ImageUrlMedium string
  Birthday time.Time `gorm:"null"`
  Gender string `gorm:"null"`
  Bio string `gorm:"type:text;null"`
  Searchable bool
  PersonID uint `gorm:"size:4"`
  Location string `gorm:"null"`
  FullName string `gorm:"size:70"`
  Nsfw bool
}

func (p *Profile) Cast(entity *federation.EntityProfile) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  var person Person
  err = db.Where("author = ?", entity.Author).First(&person).Error
  if err != nil {
    return
  }

  birthday, timeErr := time.Parse("2006-02-01", entity.Birthday)
  if timeErr == nil {
    (*p).Birthday = birthday
  }

  (*p).Author = entity.Author
  (*p).FirstName = entity.FirstName
  (*p).LastName = entity.LastName
  (*p).ImageUrl = entity.ImageUrl
  (*p).ImageUrlMedium = entity.ImageUrlMedium
  (*p).ImageUrlSmall = entity.ImageUrlSmall
  (*p).Gender = entity.Gender
  (*p).Bio = entity.Bio
  (*p).Searchable = entity.Searchable
  (*p).PersonID = person.ID
  (*p).Location = entity.Location
  (*p).FullName = entity.FirstName + " " + entity.LastName
  (*p).Nsfw = entity.Nsfw

  return
}

func (p Profile) Nickname() (nickname string) {
  nickname, _, _ = helpers.ParseAuthor(p.Author)
  return
}

func (profile *Profile) FindByPersonID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("person_id = ?", id).First(profile).Error
}
