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
  "bytes"
)

type StreamTest struct {
  GnggTestSuite
}

func (t *StreamTest) Before() {
  t.ClearDB()
  t.CreateUser()
}

func (t *StreamTest) TestPagination() {
  body := t.GET("/stream?view=public")
  t.AssertEqual(true, bytes.Contains(
    body, []byte("/stream?view=public&page=2")))

  body = t.GET("/stream?view=public&page=2")
  t.AssertEqual(true, bytes.Contains(
    body, []byte("/stream?view=public&page=3")))
  t.AssertEqual(true, bytes.Contains(
    body, []byte("/stream?view=public&page=1")))
}
