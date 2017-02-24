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
  "github.com/ganggo/federation"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Post struct {
  gorm.Model

  PersonID uint
  Public bool
  Guid string
  Type string
  Text string
  ProviderName string
  //RootGuid string
  LikesCount int
  CommentsCount int
  ResharesCount int
  InteractedAt string

  Person Person
  Comments []Comment
}

type Posts []Post

func (p *Post) Cast(entity *federation.EntityStatusMessage) (err error) {
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

  (*p).PersonID = person.ID
  (*p).Public = entity.Public
  (*p).Guid = entity.Guid
  (*p).Type = StatusMessage
  (*p).Text = entity.RawMessage
  (*p).ProviderName = entity.ProviderName

  return nil
}
