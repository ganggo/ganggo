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
  "net/http"
  "github.com/revel/revel"
  helpers "git.feneas.org/ganggo/federation/helpers"
  "git.feneas.org/ganggo/federation"
  run "github.com/revel/modules/jobs/app/jobs"
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/ganggo/app/jobs"
  "io/ioutil"
)

type Receiver struct {
  *revel.Controller
}

func (r Receiver) Public() revel.Result {
  db, err := models.OpenDatabase()
  if err != nil {
    r.Log.Error("Cannot open database", "error", err)
    return r.RenderError(err)
  }
  defer db.Close()

  request := r.Request.In.GetRaw().(*http.Request)
  content, err := ioutil.ReadAll(request.Body)
  if err != nil {
    r.Response.Status = http.StatusNotAcceptable
    return r.Render()
  }

  r.Log.Debug("received publicly", "message", string(content))

  // in case it succeeds reply with status 202
  r.Response.Status = http.StatusAccepted

  msg, err := federation.DiasporaParse(content)
  if err != nil {
    r.Log.Debug("Cannot parse content", err.Error(), err)
  } else {
    run.Now(jobs.Receiver{Message: msg})
  }
  return r.Render()
}

func (r Receiver) Private() revel.Result {
  var (
    guid string
    person models.Person
    user models.User
  )

  db, err := models.OpenDatabase()
  if err != nil {
    r.Log.Error("Cannot open database", "error", err)
    return r.RenderError(err)
  }
  defer db.Close()

  r.Params.Bind(&guid, "guid")
  r.Response.Status = http.StatusAccepted

  r.Log.Debug("received privately", "message", string(r.Params.JSON))

  err = db.Where("guid like ?", guid).First(&person).Error
  if err != nil {
    r.Log.Error("Cannot find person", "error", err)
    // diaspora will not process on StatusNotAcceptable
    return r.Render()
  }

  err = db.First(&user, person.UserID).Error
  if err != nil {
    r.Log.Error("Cannot find user", "error", err)
    r.Response.Status = http.StatusNotAcceptable
    return r.Render()
  }

  privKey, err := helpers.ParseRSAPrivateKey(
    []byte(user.SerializedPrivateKey))
  if err != nil {
    r.Log.Error(err.Error())
    r.Response.Status = http.StatusInternalServerError
    return r.Render()
  }

  msg, err := federation.DiasporaParseEncrypted(r.Params.JSON, privKey)
  if err != nil {
    r.Log.Debug("Cannot parse content", err.Error(), err)
  } else {
    run.Now(jobs.Receiver{Message: msg, Guid: guid})
  }
  return r.Render()
}
