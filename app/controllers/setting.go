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

type Setting struct {
  *revel.Controller
}

func (s Setting) Index() revel.Result {
  user, err := models.CurrentUser(s.Controller)
  if err != nil {
    s.Log.Error("Cannot fetch current user", "error", err)
    return s.RenderError(err)
  }
  s.ViewArgs["currentUser"] = user

  var tokens models.OAuthTokens
  err = tokens.FindByUserID(user.ID)
  if err != nil {
    s.Log.Error("Cannot fetch user tokens", "error", err)
    return s.RenderError(err)
  }
  s.ViewArgs["tokens"] = tokens

  return s.RenderTemplate("user/settings.html")
}
