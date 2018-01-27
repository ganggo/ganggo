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
)

func (receiver *Receiver) Contact(entity federation.EntityContact) {
  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Warn(err.Error())
    return
  }
  defer db.Close()

  revel.AppLog.Debug("Found a contact entity", "entity", entity)

  var contact models.Contact
  err = contact.Cast(&entity)
  if err != nil {
    revel.AppLog.Warn(err.Error())
    return
  }

  var oldContact models.Contact
  err = db.Where("user_id = ? AND person_id = ?",
    contact.UserID, contact.PersonID,
  ).First(&oldContact).Error
  if err == nil {
    if err = db.Model(&oldContact).Updates(
      map[string]interface{}{
        "sharing": contact.Sharing,
        "receiving": contact.Receiving,
      },
    ).Error; err != nil {
      revel.AppLog.Warn(err.Error())
      return
    }
  } else {
    err = db.Create(&contact).Error
    if err != nil {
      revel.AppLog.Warn(err.Error())
      return
    }
  }
}
