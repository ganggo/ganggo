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
  diaspora "github.com/ganggo/federation/diaspora"
)

type Photo struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Guid string `gorm:"size:191"`
  RemotePath string `gorm:"type:text"`
  Public bool
  PersonID uint
  Text string `gorm:"type:text"`
  PostID uint
  Height int
  Width int

  Person Person
}

type Photos []Photo

func (p Photos) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  for _, photo := range p {
    err = db.Create(&photo).Error
    if err != nil {
      return err
    }
  }
  return nil
}

func (p *Photos) Cast(entities diaspora.EntityPhotos) error {
  for _, entity := range entities {
    var person Person
    err := person.FindByAuthor(entity.Author)
    if err != nil {
      return err
    }

    photo := Photo{
      Guid: entity.Guid,
      Public: entity.Public,
      RemotePath: entity.RemotePhotoPath + entity.RemotePhotoName,
      Text: entity.Text,
      PersonID: person.ID,
      Height: entity.Height,
      Width: entity.Width,
    }
    *p = append(*p, photo)
  }
  return nil
}
