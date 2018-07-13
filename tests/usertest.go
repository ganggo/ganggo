package tests
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
  "net/url"
  "net/http"
  "git.feneas.org/ganggo/ganggo/app/helpers"
)

type UserTest struct {
  GnggTestSuite
}

func (t *UserTest) TestCreateUser() {
  values := url.Values{}
  for name := range helpers.UserBlacklist {
    values.Set("username", name)
    values.Set("password", "pppppp")
    values.Set("confirm", "pppppp")

    t.PostForm("/users/sign_up", values)
    t.Assertf(t.Response.StatusCode != http.StatusOK,
      "Expected status code %d for %s, got %d", http.StatusOK, name, t.Response.StatusCode)
  }
}
