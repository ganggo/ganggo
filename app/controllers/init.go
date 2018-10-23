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
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Setting{})
  revel.InterceptFunc(requiresLogin, revel.BEFORE, &Search{})
}

func redirectIfLoggedIn(c *revel.Controller) revel.Result {
  result := requiresLogin(c)
  if result == nil {
    return c.Redirect(Stream.Index)
  }
  return nil
}

func requiresLogin(c *revel.Controller) revel.Result {
  _, err := models.CurrentUser(c)
  if err != nil {
    c.Flash.Error(c.Message("flash.errors.login"))
    return c.Redirect(User.Login)
  }
  return nil
}
