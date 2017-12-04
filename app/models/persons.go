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

type Person struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Guid string `gorm:"size:191"`
  Author string `gorm:"size:191"`
  SerializedPublicKey string `gorm:"type:text"`
  UserID uint `gorm:"size:4"`
  ClosedAccount int
  FetchStatus int `gorm:"size:4"`
  PodID uint `gorm:"size:4"`

  Profile Profile `json:",omitempty"`
  Contacts Contacts `json:",omitempty"`
}

// load relations on default
func (person *Person) AfterFind() error {
  if structLoaded(person.Profile.CreatedAt) {
    return nil
  }

  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Model(person).Related(&person.Profile).Error
}

func (person *Person) FindByID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(person, id).Error
}

func (person *Person) FindByGuid(guid string) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("guid = ?", guid).First(person).Error
}

func (person *Person) FindByAuthor(author string) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("author = ?", author).First(person).Error
}
