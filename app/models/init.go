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
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
  "fmt"
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

func OpenDatabase() (*gorm.DB, error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    return db, err
  }
  db.SetLogger(helpers.AppLogWrapper{Name: "gorm"})
  db.LogMode(true)
  return db, err
}

func GetCurrentUser(token string) (user User, err error) {
  db, err := OpenDatabase()
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

func parentIsLocal(postID uint) (user User, found bool) {
  db, err := OpenDatabase()
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

// This is required since gorm.ModifyColumn only supports postgres engine
// see https://github.com/jinzhu/gorm/blob/0a51f6cdc55d1650d9ed3b4c13026cfa9133b01e/scope.go#L1142
func advancedColumnModify(s *gorm.DB, column, dataType string) {
  var format string
  var scope = s.NewScope(s.Value)

  switch DB.Driver {
    case "postgres":
      format = "ALTER TABLE %v ALTER COLUMN %v TYPE %v"
    case "mysql":
      format = "ALTER TABLE %v MODIFY %v %v"
    case "mssql":
      format = "ALTER TABLE %v ALTER COLUMN %v %v"
    default:
      revel.AppLog.Warn("Database doesn't support alter! Please do it manually",
        "driver", DB.Driver, "table", scope.QuotedTableName(),
        "column", column, "type", dataType)
      return
  }
  // modify column in scope
  scope.Raw(fmt.Sprintf(
    format, scope.QuotedTableName(),
    scope.Quote(column), dataType,
  )).Exec()
}
