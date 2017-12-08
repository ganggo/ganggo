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
  api "gopkg.in/ganggo/api.v0/app/controllers"
)

func init() {
  // redirect if logged-in
  revel.InterceptFunc(redirectIfLoggedIn, revel.BEFORE, &App{})
  // requires login
  revel.InterceptFunc(requiresHTTPLogin, revel.BEFORE, &Stream{})
  revel.InterceptFunc(requiresHTTPLogin, revel.BEFORE, &Setting{})
  revel.InterceptFunc(requiresHTTPLogin, revel.BEFORE, &Search{})
  // API
  revel.InterceptFunc(requiresTokenLogin, revel.BEFORE, &api.ApiComment{})
  revel.InterceptFunc(requiresTokenLogin, revel.BEFORE, &api.ApiLike{})
  revel.InterceptFunc(requiresTokenLogin, revel.BEFORE, &api.ApiPost{})
  revel.InterceptFunc(requiresTokenLogin, revel.BEFORE, &api.ApiPeople{})
  revel.InterceptFunc(requiresTokenLogin, revel.BEFORE, &api.ApiProfile{})
  revel.InterceptFunc(requiresTokenLogin, revel.BEFORE, &api.ApiAspect{})
  revel.InterceptFunc(requiresTokenLogin, revel.BEFORE, &api.ApiNotification{})
}

func redirectIfLoggedIn(c *revel.Controller) revel.Result {
  result := requiresHTTPLogin(c)
  if result == nil {
    return c.Redirect(Stream.Index)
  }
  return nil
}

func requiresTokenLogin(c *revel.Controller) revel.Result {
  var accessToken string

  c.Params.Bind(&accessToken, "access_token")
  if accessToken != "" {
    var token models.OAuthToken
    err := token.FindByToken(accessToken)
    if err != nil {
      c.Log.Error("Cannot find token", "error", err)
      return c.RenderError(err)
    }
    return nil
  }
  // fallback to http authentication
  return requiresHTTPLogin(c)
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
