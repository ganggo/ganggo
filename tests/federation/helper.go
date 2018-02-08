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
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/gorm.v2"
  _ "gopkg.in/ganggo/gorm.v2/dialects/postgres"
  "os"
  "fmt"
)

var userRelations bool = false
var ciDatabases = [2]string{"d1", "d2"}

const (
  username = "ganggo"
  handle = username + "@localhost:9000"
  password = "pppppp"
)

type TokenJSON struct {
  Token string `json:"token"`
}

type FederationSuite struct {
  testing.TestSuite
}

func (t* FederationSuite) DB(databaseName string) (*gorm.DB, error) {
  url := fmt.Sprintf("user=postgres dbname=%s sslmode=disable", databaseName)
  db, err := gorm.Open("postgres", url)
  if err != nil {
    return db, err
  }
  return db, err
}

func (t *FederationSuite) SetupUserRelations() error {
  if userRelations { return nil }

  t.CreateUser()

  if t.CI() {
    var (
      alice models.Person
      bob models.Person
      aspect models.Aspect
    )
    values := url.Values{}
    // search for alice
    values.Set("handle", "d1@localhost:3000")
    result := t.POST("/api/v0/search", values)
    err := json.Unmarshal(result, &alice)
    if err != nil {
      return err
    }
    values = url.Values{}
    // search for bob
    values.Set("handle", "d2@localhost:3001")
    result = t.POST("/api/v0/search", values)
    err = json.Unmarshal(result, &bob)
    if err != nil {
      return err
    }
    values = url.Values{}
    // create an aspect
    values.Set("aspect_name", "testsuite")
    result = t.POST("/api/v0/aspects", values)
    err = json.Unmarshal(result, &aspect)
    if err != nil {
      return err
    }
    // add both to the same aspect
    t.POST(fmt.Sprintf(
      "/api/v0/people/%d/aspects/%d", alice.ID, aspect.ID,
    ), url.Values{})
    t.POST(fmt.Sprintf(
      "/api/v0/people/%d/aspects/%d", bob.ID, aspect.ID,
    ), url.Values{})
  }
  userRelations = true
  return nil
}

func (t *FederationSuite) CreateUser() {
  values := url.Values{}
  values.Set("username", username)
  values.Set("password", password)
  values.Set("confirm", password)

  t.PostForm("/users/sign_up", values)
}

func (t *FederationSuite) AccessToken() string {
  values := url.Values{}
  values.Set("username", username)
  values.Set("password", password)
  values.Set("grant_type", "password")
  values.Set("client_id", "testsuite")

  t.PostForm("/api/v0/oauth/tokens", values)

  var token TokenJSON
  err := json.Unmarshal(t.ResponseBody, &token)
  if err != nil {
    panic("Cannot unmarshal api token: " + err.Error())
  }
  return token.Token
}

func (t *FederationSuite) POST(path string, values url.Values) []byte {
  req := t.PostFormCustom(t.BaseUrl() + path, values)
  req.Header.Set("access_token", t.AccessToken())
  req.Send()
  return t.ResponseBody
}

func (t *FederationSuite) CI() bool {
  return os.Getenv("CI") != ""
}
