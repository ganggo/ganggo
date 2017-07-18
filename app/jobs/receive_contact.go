package jobs
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
  "gopkg.in/ganggo/ganggo.v0/app/models"
  federation "gopkg.in/ganggo/federation.v0"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

func (r *Receiver) Contact(entity federation.EntityContact) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  revel.TRACE.Println("Found a contact entity", entity)

  var contact models.Contact
  err = contact.Cast(&entity)
  if err != nil {
    revel.WARN.Println(err)
    return
  }

  var oldContact models.Contact
  if err = db.Where(
    "user_id = ? AND person_id = ?",
    contact.UserID, contact.PersonID,
  ).First(&oldContact).Error; err == nil {
    err = db.Model(&oldContact).Updates(contact).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
  } else {
    err = db.Create(&contact).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
  }
}
