package helpers
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

import "github.com/revel/revel"

func IsLocalHandle(handle string) bool {
  revel.Config.SetSection("ganggo")
  localhost, found := revel.Config.String("address")
  if !found {
    panic("No server address configured")
  }

  _, host, err := ParseAuthor(handle)
  if err != nil {
    panic("Cannot parse diaspora handle")
  }

  if host == localhost {
    return true
  }
  return false
}
