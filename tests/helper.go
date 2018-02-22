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

import (
  "net/url"
  "encoding/json"
  "github.com/revel/revel/testing"
  "github.com/ganggo/ganggo/app/models"
)

const (
  username = "ganggo"
  handle = username + "@localhost:9000"
  password = "pppppp"
)

type TokenJSON struct {
  Token string `json:"token"`
}

type GnggTestSuite struct {
  testing.TestSuite
}

func (t *GnggTestSuite) ClearDB() {
  // re-create ganggo database
  db, err := models.OpenDatabase()
  t.Assertf(err == nil, "Expected nil, got '%+v'", err)
  defer db.Close()

  tables := []interface{}{
    models.Aspect{},
    models.AspectMembership{},
    models.AspectVisibility{},
    models.Comment{},
    models.CommentSignature{},
    models.Contact{},
    models.Like{},
    models.LikeSignature{},
    models.Notification{},
    models.OAuthToken{},
    models.Person{},
    models.Photo{},
    models.Pod{},
    models.Post{},
    models.Profile{},
    models.Session{},
    models.Shareable{},
    models.SignatureOrder{},
    models.Tag{},
    models.ShareableTagging{},
    models.UserTagging{},
    models.User{},
    models.UserStream{},
  }

  for i, table := range tables {
    err = db.Unscoped().Delete(table).Error
    t.Assertf(err == nil, "#%d: Expected nil, got '%+v'", i, err)
  }
}

func (t *GnggTestSuite) CreateUser() {
  values := url.Values{}
  values.Set("username", username)
  values.Set("password", password)
  values.Set("confirm", password)

  t.PostForm("/users/sign_up", values)
  t.AssertOk()
}

func (t *GnggTestSuite) UserID(token string) uint {
  db, err := models.OpenDatabase()
  t.Assertf(err == nil, "Expected nil, got '%+v'", err)
  defer db.Close()

  var oauth = models.OAuthToken{}
  err = db.Where("token = ?", token).Find(&oauth).Error
  t.Assertf(err == nil, "Expected nil, got '%+v'", err)
  return oauth.UserID
}

func (t *GnggTestSuite) AccessToken() string {
  values := url.Values{}
  values.Set("username", username)
  values.Set("password", password)
  values.Set("grant_type", "password")
  values.Set("client_id", "testsuite")

  t.PostForm("/api/v0/oauth/tokens", values)

  var token TokenJSON
  err := json.Unmarshal(t.ResponseBody, &token)
  t.Assertf(err == nil, "Expected nil, got '%+v'", err)
  return token.Token
}

func (t *GnggTestSuite) POST(path string, values url.Values) []byte {
  req := t.PostFormCustom(t.BaseUrl() + path, values)
  req.Header.Set("access_token", t.AccessToken())
  req.Send()
  return t.ResponseBody
}

func (t *GnggTestSuite) GET(path string) []byte {
  req := t.GetCustom(t.BaseUrl() + path)
  req.Header.Set("access_token", t.AccessToken())
  req.Send()
  return t.ResponseBody
}
