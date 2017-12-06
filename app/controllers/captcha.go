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
  "github.com/dchest/captcha"
  "net/http"
  "path"
  "bytes"
)

type Captcha struct {
  *revel.Controller
}

type DisplayImageByName string

func (name DisplayImageByName) Apply(req *revel.Request, resp *revel.Response) {
  var content bytes.Buffer
  status := http.StatusOK

  _, file := path.Split(string(name))
  ext := path.Ext(file)
  id := file[:len(file)-len(ext)]

  err := captcha.WriteImage(&content, id, 250, 100)
  if err != nil {
    revel.AppLog.Error("Cannot write captcha image", "error", err)
    status = http.StatusInternalServerError
  }
  resp.WriteHeader(status, "image/png")
  resp.GetWriter().Write(content.Bytes())
}

func (c Captcha) Index(name string) revel.Result {
  return DisplayImageByName(name)
}
