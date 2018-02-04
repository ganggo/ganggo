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
)

type Profile struct {
  *revel.Controller
}

func (p Profile) Index(guid string) revel.Result {
  return p.IndexPagination(guid, 0)
}

func (p Profile) IndexPagination(guid string, page uint) revel.Result {
  var (
    offset uint = ((page - 1) * 10)
    posts models.Posts
    person models.Person
  )

  err := person.FindByGuid(guid)
  if err != nil {
    p.Log.Error("Cannot find person", "error", err)
    return p.RenderError(err)
  }

  user, err := models.CurrentUser(p.Controller)
  if err == nil {
    p.ViewArgs["currentUser"] = user
  }

  err = posts.FindAllByUserAndPersonID(user, person.ID, offset)
  if err != nil {
    p.Log.Error("Cannot find posts", "error", err)
    return p.RenderError(err)
  }

  p.ViewArgs["posts"] = posts
  p.ViewArgs["person"] = person
  if page <= 0 { page = 1 }
  p.ViewArgs["page"] = page

  return p.RenderTemplate("profile/index.html")
}
