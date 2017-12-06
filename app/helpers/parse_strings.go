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


// Will contain following parts
//  Index Match
//  0 @{ganggo@localhost; ganggo@localhost}
//  1 ganggo@localhost
//  2 ganggo
//  3 localhost
func ParseMentions(text string) [][]string {
  r := regexp.MustCompile(`@\{\s*([^;]*?)[;\s]*([^@;\s]+?)@([^@;\s]+?)\s*\}`)
  return r.FindAllStringSubmatch(text, -1)
}

func ParseTags(text string) [][]string {
  r := regexp.MustCompile(`#([\w\d]{2,})`)
  return r.FindAllStringSubmatch(text, -1)
}

func ParseAuthor(handle string) (string, string, error) {
  parts, err := parseStringHelper(handle, `^(.+?)@(.+?)$`, 2)
  if err != nil {
    return "", "", err
  }
  return parts[1], parts[2], nil
}

func ParseWebfingerHandle(handle string) (string, error) {
  parts, err := parseStringHelper(handle, `^acct:(.+?)@.+?$`, 1)
  if err != nil {
    return "", err
  }
  return parts[1], nil
}

func parseStringHelper(line, regex string, max int) (parts []string, err error) {
  r := regexp.MustCompile(regex)
  parts = r.FindStringSubmatch(line)
  if len(parts) < max {
    err = errors.New("Cannot extract " + regex + " from " + line)
    return
  }
  return
}
