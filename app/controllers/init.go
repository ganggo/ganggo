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
  "git.feneas.org/ganggo/ganggo/app/models"
)

func init() {
  // redirect if logged-in
  revel.InterceptFunc(redirectIfLoggedIn, revel.BEFORE, &App{})
  // requires login
  revel.InterceptFunc(requiresHTTPLogin, revel.BEFORE, &Setting{})
  revel.InterceptFunc(requiresHTTPLogin, revel.BEFORE, &Search{})
}

func redirectIfLoggedIn(c *revel.Controller) revel.Result {
  result := requiresHTTPLogin(c)
  if result == nil {
    return c.Redirect(Stream.Index)
  }
  return nil
}

func requiresHTTPLogin(c *revel.Controller) revel.Result {
  var session models.Session

  db, err := models.OpenDatabase()
  if err != nil {
    c.Log.Error("Cannot open database", "error", err)
    return c.RenderError(err)
  }
  defer db.Close()

  err = db.Where("token = ?", c.Session["TOKEN"]).First(&session).Error
  if err != nil {
    c.Flash.Error("Please log in first")
    return c.Redirect(App.Index)
  }
  return nil
}
