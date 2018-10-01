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
  "fmt"
  "crypto/rand"
  "crypto/md5"
)

func Uuid() (string, error) {
  b, err := randomBytes(16)
  if err != nil {
    return "", err
  }
  return fmt.Sprintf("%x", b), nil
}

func UuidFromSalt(salt string) string {
  return fmt.Sprintf("%x", md5.Sum([]byte(salt)))
}

func Token() (string, error) {
  b, err := randomBytes(32)
  if err != nil {
    return "", err
  }
  return fmt.Sprintf("%x", b), nil
}

func randomBytes(length int) (b []byte, err error) {
  b = make([]byte, length)
  _, err = rand.Read(b)
  return
}
