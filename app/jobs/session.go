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
  "time"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
)

type Session struct {}

// Run will clean-up all sessions older then two days
func (s Session) Run() {
  // from; 1970-01-01 00:00:00 +0000 UTC
  from, err := time.Parse("2006", "1970")
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // to; now() - two day
  to := time.Now().AddDate(0, 0, -2)

  var sessions models.Sessions
  err = sessions.FindByTimeRange(from, to)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  err = sessions.Delete()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
}
