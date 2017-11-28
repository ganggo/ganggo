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

type Post struct {
  *revel.Controller
}

func (p Post) Index(guid string) revel.Result {
  var post models.Post

  err := post.FindByGuid(guid)
  if err != nil {
    p.Response.Status = http.StatusInternalServerError
    revel.WARN.Println(err)
  }

  user, err := models.GetCurrentUser(p.Session["TOKEN"])
  if err == nil {
    p.ViewArgs["currentUser"] = user
  }


  return p.Render(post)
}
