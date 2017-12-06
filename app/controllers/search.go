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
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  "gopkg.in/ganggo/ganggo.v0/app/jobs"
)

type Search struct {
  *revel.Controller
}

func (s Search) Create(text string) revel.Result {
  _, _, err := helpers.ParseAuthor(text)
  if err != nil {
    return s.NotFound(err.Error())
  }

  fetchAuthor := jobs.FetchAuthor{Author: text}
  fetchAuthor.Run()
  if fetchAuthor.Err != nil {
    s.Log.Error("Cannot fetch author", "error", fetchAuthor.Err)
    return s.RenderError(fetchAuthor.Err)
  }

  guid := fetchAuthor.Person.Guid
  return s.Redirect("/profiles/%s", guid)
}
