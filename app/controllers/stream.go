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

type Stream struct {
  *revel.Controller
}

func (s Stream) Index() revel.Result {
  return s.IndexPagination(0)
}

func (s Stream) IndexPagination(page int) revel.Result {
  var posts models.Posts
  var offset int = ((page - 1) * 10)

  user, err := models.CurrentUser(s.Controller)
  if err != nil {
    s.Log.Error("Cannot fetch current user", "error", err)
    return s.RenderError(err)
  }

  err = posts.FindAllPrivate(user.ID, offset)
  if err != nil {
    s.Log.Error("Cannot find posts", "error", err)
    return s.RenderError(err)
  }

  s.ViewArgs["title"] = revel.MessageFunc(
    s.Request.Locale, "stream.title",
  )
  s.ViewArgs["currentUser"] = user
  if page == 0 { page = 1 }
  s.ViewArgs["page"] = page
  s.ViewArgs["posts"] = posts

  return s.RenderTemplate("stream/index.html")
}
