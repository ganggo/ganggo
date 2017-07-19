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
  "github.com/revel/revel"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Database struct {
  Driver string
  Url string
}

const (
  Reshare = "Reshare"
  StatusMessage = "StatusMessage"
  ShareablePost = "Post"
)

var DB Database

func parentIsLocal(postID uint) (user User, found bool) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  var post Post
  // XXX here we assume every comment is related to post
  // that could be a problem in respect of private messages
  err = db.First(&post, postID).Error
  if err != nil {
    return
  }
  db.Model(&post).Related(&post.Person, "Person")

  if post.Person.UserID > 0 {
    err = db.First(&user, post.Person.UserID).Error
    if err != nil {
      return
    }
    db.Model(&user).Related(&user.Person, "Person")
    found = true
    return
  }
  return
}

func GetCurrentUser(token string) (user User, err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return user, err
  }
  defer db.Close()

  var session Session
  err = db.Where("token = ?", token).First(&session).Error
  if err != nil {
    revel.ERROR.Println(err)
    return user, err
  }

  err = db.First(&user, session.UserID).Error
  if err != nil {
    revel.ERROR.Println(err)
    return user, err
  }
  db.Model(&user).Related(&user.Person, "Person")
  db.Model(&user).Related(&user.Aspects)
  return
}
