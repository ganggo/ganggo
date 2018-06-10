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
  "github.com/ganggo/ganggo/app/models"
  "github.com/ganggo/federation"
)

func (receiver *Receiver) Profile(profile federation.MessageProfile) {
  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  var profileModel models.Profile
  err = profileModel.FindByAuthor(profile.Author())
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  err = profileModel.Cast(profile)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  err = db.Save(&profileModel).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
}
