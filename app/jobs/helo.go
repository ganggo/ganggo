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
  federation "gopkg.in/ganggo/federation.v0"
  run "github.com/revel/modules/jobs/app/jobs"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
)

type Helo struct {}

// Run tries to start sharing with unknown servers
func (h Helo) Run() {
  var pods models.Pods
  err := pods.FindByHelo(false)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  var user models.User
  err = user.FindByUsername("hq")
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  for _, pod := range pods {
    var person models.Person
    err = person.FirstByPod(pod)
    if err != nil {
      revel.AppLog.Error(err.Error())
      continue
    }

    //XXX x-social-relay

    run.Now(Dispatcher{
      User: user,
      Message: federation.EntityContact{
        Author: user.Person.Author,
        Recipient: person.Author,
        Sharing: true,
        Following: true,
      },
    })
  }
}
