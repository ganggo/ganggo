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

func (s Search) Create() revel.Result {
  var searchText string
  s.Params.Bind(&searchText, "search")

  _, _, err := helpers.ParseDiasporaHandle(searchText)
  if err != nil {
    return s.Redirect(Stream.Index)
  }

  fetchAuthor := jobs.FetchAuthor{
    Author: searchText,
  }
  fetchAuthor.Run()
  if fetchAuthor.Err != nil {
    revel.WARN.Println(fetchAuthor.Err)
    return s.Redirect(Stream.Index)
  }

  guid := (*fetchAuthor.Person).Guid
  return s.Redirect("/profiles/%s", guid)
}
