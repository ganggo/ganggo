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

func (p Profile) IndexPagination(guid string, page int) revel.Result {
  var (
    offset int = ((page - 1) * 10)
    posts models.Posts
    person models.Person
  )

  err := person.FindByGuid(guid)
  if err != nil {
    revel.ERROR.Println(err)
    return p.Redirect(Stream.Index)
  }

  err = posts.FindAllByPersonID(person.ID, offset)
  if err != nil {
    revel.ERROR.Println(err)
    return p.Redirect(Stream.Index)
  }

  p.ViewArgs["posts"] = posts
  p.ViewArgs["person"] = person
  p.ViewArgs["page"] = page

  user, err := models.GetCurrentUser(p.Session["TOKEN"])
  if err == nil {
    p.ViewArgs["currentUser"] = user
  }

  return p.RenderTemplate("profile/index.html")
}
