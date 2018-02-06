package controllers
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
  api "gopkg.in/ganggo/api.v0/app/controllers"
)

type People struct {
  *revel.Controller
}

func (p People) IndexStream(guid, fields string, offset uint) revel.Result {
  controller := p.Controller
  controller.Params.Add("format", "diaspora")
  userStream := api.ApiUserStream{
    api.ApiHelper{controller, models.User{}}}
  return userStream.ShowPersonStream(guid, fields, offset)
}
