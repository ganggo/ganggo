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
  "strings"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "git.feneas.org/ganggo/ganggo/app/jobs"
  "git.feneas.org/ganggo/ganggo/app/models"
)

type Search struct {
  *revel.Controller
}

func (s Search) Index(text string) revel.Result {
  return s.IndexPagination(text, 0)
}

func (s Search) IndexPagination(text string, page uint) revel.Result {
  var offset uint = ((page - 1) * 10)
  text = strings.Replace(text, "'", "", -1)

  user, err := models.CurrentUser(s.Controller)
  if err != nil {
    s.Log.Error("Cannot fetch current user", "error", err)
    return s.RenderError(err)
  }
  s.ViewArgs["currentUser"] = user

  var posts models.Posts
  err = posts.FindAllByUserAndText(user, text, offset)
  if err != nil {
    s.Log.Error("Cannot find posts", "error", err)
    return s.RenderError(err)
  }
  s.ViewArgs["posts"] = posts
  if page == 0 { page = 1 }
  s.ViewArgs["page"] = page
  s.ViewArgs["searchQuery"] = text

  return s.RenderTemplate("search/index.html")
}

func (s Search) Create(search string) revel.Result {
  _, _, err := helpers.ParseAuthor(search)
  if err != nil {
    s.Log.Debug("Cannot parse handle author", "error", err)
    return s.Redirect("/search/%s", search)
  }

  fetchAuthor := jobs.FetchAuthor{Author: search}
  fetchAuthor.Run()
  if fetchAuthor.Err != nil {
    s.Log.Error("Cannot fetch author", "error", fetchAuthor.Err)
    return s.RenderError(fetchAuthor.Err)
  }
  guid := fetchAuthor.Person.Guid

  return s.Redirect("/profiles/%s", guid)
}
