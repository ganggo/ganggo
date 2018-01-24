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

import "github.com/revel/revel/testing"

type AppTest struct {
  testing.TestSuite
}

func (t *AppTest) Before() {
  println("Set up")
}

func (t *AppTest) TestThatIndexPageWorks() {
  t.Get("/")
  t.AssertOk()
  t.AssertContentType("text/html; charset=utf-8")
}

func (t *AppTest) After() {
  println("Tear down")
}
