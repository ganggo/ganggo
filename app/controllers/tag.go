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
  "net/http"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
)

type Tag struct {
  *revel.Controller
}

func (t Tag) Index(name string) revel.Result {
  var posts models.Posts

  user, err := models.GetCurrentUser(t.Session["TOKEN"])
  if err == nil {
    t.ViewArgs["currentUser"] = user
  }

  err = posts.FindByTagName(name, user, 0)
  if err != nil {
    t.Response.Status = http.StatusInternalServerError
    revel.WARN.Println(err)
  }

  return t.Render(posts)
}
