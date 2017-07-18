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
  "golang.org/x/crypto/bcrypt"
  "github.com/revel/revel"
  "github.com/dchest/captcha"

  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"

  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
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

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return u.Render()
  }
  defer db.Close()

  u.Params.Bind(&username, "username")
  u.Params.Bind(&password, "password")
  u.Params.Bind(&verify, "confirm")

  u.Params.Bind(&captchaID, "captchaID")
  u.Params.Bind(&captchaValue, "captchaValue")

  if !captcha.VerifyString(captchaID, captchaValue) {
    u.Flash.Error("Captcha was not correct!")
    return u.Redirect(User.Index)
  }

  if !db.Where("username = ?", username).First(&user).RecordNotFound() {
    u.Flash.Error("Username already exists!")
    return u.Redirect(User.Index)
  }

  if password == "" || password != verify {
    u.Flash.Error("Password was empty or didn't match!")
    return u.Redirect(User.Index)
  }

  if len(password) < 4 {
    u.Flash.Error("Password length should be greater then four!")
    return u.Redirect(User.Index)
  }

  // generate priv,pub key
  privKey, err := rsa.GenerateKey(rand.Reader, 2048)
  if err != nil {
    revel.ERROR.Println(err)
    u.Response.Status = http.StatusInternalServerError
    return u.Redirect(User.Index)
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
    revel.ERROR.Println(err)
    u.Response.Status = http.StatusInternalServerError
    return u.Redirect(User.Index)
  }
  pubBlock := pem.Block{
    Type: "PUBLIC KEY",
    Bytes: pub,
  }
  pubEncoded := pem.EncodeToMemory(&pubBlock)

  guid, err := helpers.Uuid()
  if err != nil {
    revel.ERROR.Println(err)
    u.Response.Status = http.StatusInternalServerError
    return u.Redirect(User.Index)
  }

  passwordEncoded, err := bcrypt.GenerateFromPassword([]byte(password), -1)
  if err != nil {
    revel.ERROR.Println(err)
    u.Response.Status = http.StatusInternalServerError
    return u.Redirect(User.Index)
  }

  revel.Config.SetSection("ganggo")
  host, found := revel.Config.String("address")
  if !found {
    revel.ERROR.Println("No server address configured")
    u.Response.Status = http.StatusInternalServerError
    return u.Redirect(User.Index)
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
    u.Flash.Error("Something went wrong :(")
    revel.ERROR.Println(err)
    return u.Redirect(User.Index)
  } else {
    u.Flash.Success("The user was successfully created! Please login")
  }
  return u.Redirect(App.Index)
}

func (u User) Logout() revel.Result {
  var session models.Session

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return u.Render()
  }
  defer db.Close()

  err = db.Where("token = ?", u.Session["TOKEN"]).First(&session).Error
  if err != nil {
    u.Response.Status = http.StatusInternalServerError
    revel.ERROR.Println(err)
    return u.Render()
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
  u.Params.Bind(&username, "username")
  u.Params.Bind(&password, "password")

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return u.Render()
  }
  defer db.Close()

  err = db.Where("username = ?", username).First(&user).Error
  if err != nil {
    // TODO flash message not found
    revel.TRACE.Println(err)
    return u.Redirect(App.Index)
  }

  if !checkHash(password, user.EncryptedPassword) {
    revel.TRACE.Println("Login failed for user " + username)
    return u.Redirect(App.Index)
  }
  revel.TRACE.Println("Login successful for user " + username)

  token, err := helpers.Uuid()
  if err != nil {
    u.Response.Status = http.StatusInternalServerError
    revel.TRACE.Println(err)
    return u.Redirect(App.Index)
  }
  session.UserID = user.ID
  session.Token = token
  u.Session["TOKEN"] = session.Token

  err = db.Create(&session).Error
  if err != nil {
    u.Response.Status = http.StatusInternalServerError
    revel.TRACE.Println(err)
    delete(u.Session, "TOKEN")
    return u.Redirect(App.Index)
  }
  // TODO flash message
  return u.Redirect(Stream.Index)
}

func checkHash(password, dbPassword string) bool {
  err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
  if err !=  nil {
    return false
  }
  return true
}
