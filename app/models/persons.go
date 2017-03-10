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
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Person struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  Guid string
  DiasporaHandle string
  SerializedPublicKey string `gorm:"type:text"`
  UserID uint `gorm:"size:4"`
  ClosedAccount int
  FetchStatus int `gorm:"size:4"`
  PodID uint `gorm:"size:4"`

  Profile Profile `json:",omitempty"`
  Contacts []Contact `json:",omitempty"`
}

func (person *Person) FindByID(id uint) (err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(person, id).Error
}
