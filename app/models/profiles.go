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
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Profile struct {
  gorm.Model

  DiasporaHandle string
  FirstName string `sql:"DEFAULT:NULL"`
  LastName string `sql:"DEFAULT:NULL"`
  ImageUrl string
  ImageUrlSmall string
  ImageUrlMedium string
  Birthday time.Time `sql:"DEFAULT:NULL"`
  Gender string `sql:"DEFAULT:NULL"`
  Bio string `sql:"DEFAULT:NULL"`
  Searchable bool
  PersonID uint
  Location string `sql:"DEFAULT:NULL"`
  FullName string
  Nsfw bool
}

func (p *Profile) Cast(entity *federation.EntityProfile) (err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return
  }
  defer db.Close()

  var person Person
  err = db.Where("diaspora_handle = ?", entity.DiasporaHandle).First(&person).Error
  if err != nil {
    return
  }

  birthday, timeErr := time.Parse("2006-02-01", entity.Birthday)
  if timeErr == nil {
    (*p).Birthday = birthday
  }

  (*p).DiasporaHandle = entity.DiasporaHandle
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
  nickname, _, _ = helpers.ParseDiasporaHandle(p.DiasporaHandle)
  return
}
