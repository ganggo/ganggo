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

import (
  "regexp"
  "errors"
)

func ParseAuthor(handle string) (string, string, error) {
  parts, err := _ParseStringHelper(handle, `^(.+?)@(.+?)$`, 2)
  if err != nil {
    return "", "", err
  }
  return parts[1], parts[2], nil
}

func ParseWebfingerHandle(handle string) (string, error) {
  parts, err := _ParseStringHelper(handle, `^acct:(.+?)@.+?$`, 1)
  if err != nil {
    return "", err
  }
  return parts[1], nil
}

func _ParseStringHelper(line, regex string, max int) (parts []string, err error) {
  r := regexp.MustCompile(regex)
  parts = r.FindStringSubmatch(line)
  if len(parts) < max {
    err = errors.New("Cannot extract " + regex + " from " + line)
    return
  }
  return
}
