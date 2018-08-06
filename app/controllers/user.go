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
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "net/http"
)

type User struct {
  *revel.Controller
}

func (u User) Index() revel.Result {
  return u.Render()
}

func (u User) Create() revel.Result {
  var user models.User
  var username, password, verify string
  var captchaID, captchaValue string

  db, err := models.OpenDatabase()
  if err != nil {
    revel.WARN.Println(err)
    u.Log.Error("Cannot open database", "error", err)
    return u.RenderError(err)
  }
  defer db.Close()

  u.Params.Bind(&username, "username")
  u.Params.Bind(&password, "password")
  u.Params.Bind(&verify, "confirm")

  u.Params.Bind(&captchaID, "captchaID")
  u.Params.Bind(&captchaValue, "captchaValue")

  if !revel.DevMode && !captcha.VerifyString(captchaID, captchaValue) {
    u.Flash.Error(u.Message("flash.errors.captcha"))
    return u.Redirect(User.Index)
  }

  if _, exists := helpers.UserBlacklist[username]; exists {
    u.Flash.Error(u.Message("flash.errors.username"))
    return u.Redirect(User.Index)
  }

  if !db.Where("username = ?", username).First(&user).RecordNotFound() {
    u.Flash.Error(u.Message("flash.errors.username"))
    return u.Redirect(User.Index)
  }

  if password == "" || password != verify {
    u.Flash.Error(u.Message("flash.errors.password_empty"))
    return u.Redirect(User.Index)
  }

  if len(password) < 4 {
    u.Flash.Error(u.Message("flash.errors.password_length"))
    return u.Redirect(User.Index)
  }

  // build user struct
  user = models.User{
    Password: password,
    Username: username,
    Person: models.Person {
      Profile: models.Profile{
        Searchable: true,
        ImageUrl: "/public/img/avatar.png",
      },
    },
  }

  err = db.Create(&user).Error
  if err != nil {
    u.Log.Error("Cannot create user", "error", err)
    u.Response.Status = http.StatusInternalServerError
    return u.RenderError(err)
  }
  u.Flash.Success(u.Message("flash.success.registration"))
  return u.Redirect(User.Login)
}

func (u User) Logout() revel.Result {
  var session models.Session

  db, err := models.OpenDatabase()
  if err != nil {
    u.Log.Error("Cannot open database", "error", err)
    u.Response.Status = http.StatusInternalServerError
    return u.RenderError(err)
  }
  defer db.Close()

  err = db.Where("token = ?", u.Session["TOKEN"]).First(&session).Error
  if err != nil {
    u.Log.Error("Cannot find session", "error", err)
    u.Response.Status = http.StatusInternalServerError
    return u.RenderError(err)
  }
  db.Delete(&session)
  delete(u.Session, "TOKEN")
  return u.Redirect(App.Index)
}

func (u User) Login() revel.Result {
  var (
    username string
    password string
    user models.User
    session models.Session
  )

  // render the login screen on GET
  if u.Request.Method == "GET" {
    return u.Render()
  }

  u.Params.Bind(&username, "username")
  u.Params.Bind(&password, "password")

  db, err := models.OpenDatabase()
  if err != nil {
    u.Log.Error("Cannot open database", "error", err)
    u.Response.Status = http.StatusInternalServerError
    return u.RenderError(err)
  }
  defer db.Close()

  err = db.Where("username = ?", username).First(&user).Error
  if err != nil {
    u.Flash.Error(u.Message(
      "flash.errors.username_not_found",  username))
    return u.Redirect(User.Login)
  }

  if !helpers.CheckHash(password, user.EncryptedPassword) {
    u.Flash.Error(u.Message("flash.errors.login_failed"))
    return u.Redirect(User.Login)
  }

  token, err := helpers.Uuid()
  if err != nil {
    u.Log.Error("Cannot generate UUID", "error", err)
    u.Response.Status = http.StatusInternalServerError
    return u.RenderError(err)
  }
  session.UserID = user.ID
  session.Token = token

  err = db.Create(&session).Error
  if err != nil {
    u.Log.Error("Cannot create session", "error", err)
    u.Response.Status = http.StatusInternalServerError
    return u.RenderError(err)
  }
  u.Session["TOKEN"] = session.Token

  u.Flash.Success(u.Message("flash.success.login"))
  return u.Redirect(Stream.Index)
}
