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
  "errors"
  "golang.org/x/crypto/bcrypt"
  "github.com/revel/revel"
  "github.com/dchest/captcha"

  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"

  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
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
    u.Flash.Error(revel.MessageFunc(
      u.Request.Locale,
      "flash.errors.captcha",
    ))
    return u.Redirect(User.Index)
  }

  if !db.Where("username = ?", username).First(&user).RecordNotFound() {
    u.Flash.Error(revel.MessageFunc(
      u.Request.Locale,
      "flash.errors.username",
    ))
    return u.Redirect(User.Index)
  }

  if password == "" || password != verify {
    u.Flash.Error(revel.MessageFunc(
      u.Request.Locale,
      "flash.errors.password_empty",
    ))
    return u.Redirect(User.Index)
  }

  if len(password) < 4 {
    u.Flash.Error(revel.MessageFunc(
      u.Request.Locale,
      "flash.errors.password_length",
    ))
    return u.Redirect(User.Index)
  }

  // generate priv,pub key
  privKey, err := rsa.GenerateKey(rand.Reader, 2048)
  if err != nil {
    u.Log.Error("Cannot generate RSA key", "error", err)
    return u.RenderError(err)
  }
  // private key
  key := x509.MarshalPKCS1PrivateKey(privKey)
  block := pem.Block{
    Type: "PRIVATE KEY",
    Bytes: key,
  }
  keyEncoded := pem.EncodeToMemory(&block)
  // public key
  pub, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
  if err != nil {
    u.Log.Error("Cannot extract public key", "error", err)
    return u.RenderError(err)
  }
  pubBlock := pem.Block{
    Type: "PUBLIC KEY",
    Bytes: pub,
  }
  pubEncoded := pem.EncodeToMemory(&pubBlock)

  guid, err := helpers.Uuid()
  if err != nil {
    u.Log.Error("Cannot generate UUID", "error", err)
    return u.RenderError(err)
  }

  passwordEncoded, err := bcrypt.GenerateFromPassword([]byte(password), -1)
  if err != nil {
    u.Log.Error("Cannot encode password", "error", err)
    return u.RenderError(err)
  }

  revel.Config.SetSection("ganggo")
  host, found := revel.Config.String("address")
  if !found {
    err = errors.New("No server address configured")
    u.Log.Error("", "error", err)
    return u.RenderError(err)
  }

  // build user struct
  user = models.User{
    Username: username,
    SerializedPrivateKey: string(keyEncoded),
    EncryptedPassword: string(passwordEncoded),
    Person: models.Person {
      Guid: guid,
      Author: username + "@" + host,
      SerializedPublicKey: string(pubEncoded),
      Profile: models.Profile{
        Author: username + "@" + host,
        Searchable: true,
        ImageUrl: "/public/img/avatar.png",
      },
    },
  }
  err = db.Create(&user).Error
  if err != nil {
    u.Log.Error("Cannot create user", "error", err)
    return u.RenderError(err)
  } else {
    u.Flash.Success(revel.MessageFunc(
      u.Request.Locale,
      "flash.success.registration",
    ))
  }
  return u.Redirect(User.Login)
}

func (u User) Logout() revel.Result {
  var session models.Session

  db, err := models.OpenDatabase()
  if err != nil {
    u.Log.Error("Cannot open database", "error", err)
    return u.RenderError(err)
  }
  defer db.Close()

  err = db.Where("token = ?", u.Session["TOKEN"]).First(&session).Error
  if err != nil {
    u.Log.Error("Cannot find session", "error", err)
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
    return u.RenderError(err)
  }
  defer db.Close()

  err = db.Where("username = ?", username).First(&user).Error
  if err != nil {
    u.Flash.Error(revel.MessageFunc(
      u.Request.Locale,
      "flash.errors.username_not_found",
      username,
    ))
    return u.Redirect(User.Login)
  }

  if !helpers.CheckHash(password, user.EncryptedPassword) {
    u.Flash.Error(revel.MessageFunc(
      u.Request.Locale,
      "flash.errors.login_failed",
    ))
    return u.Redirect(User.Login)
  }

  token, err := helpers.Uuid()
  if err != nil {
    u.Log.Error("Cannot generate UUID", "error", err)
    return u.RenderError(err)
  }
  session.UserID = user.ID
  session.Token = token

  err = db.Create(&session).Error
  if err != nil {
    u.Log.Error("Cannot create session", "error", err)
    return u.RenderError(err)
  }
  u.Session["TOKEN"] = session.Token

  u.Flash.Success(revel.MessageFunc(
    u.Request.Locale,
    "flash.success.login",
  ))
  return u.Redirect(Stream.Index)
}
