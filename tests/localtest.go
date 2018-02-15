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
  "time"
  "encoding/json"
  "gopkg.in/ganggo/gorm.v2"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "fmt"
)

type FederationLocalTest struct {
  FederationSuite
}

func (t *FederationLocalTest) Before() {
  println("Setup user relations and access token")
  t.AssertEqual(nil, t.SetupUserRelations())
}

func (t *FederationLocalTest) TestLocal() {
  db, err := models.OpenDatabase()
  t.AssertEqual(nil, err)
  defer db.Close()

  msg := "FederationLocalTest"
  values := url.Values{}
  values.Set("aspectID", "0") // public post
  values.Set("post", msg)

  t.POST("/api/v0/posts", values)
  var count int
  var post models.Post
  err = db.Where("text = ?", msg).Find(
    &post).Count(&count).Error
  t.AssertEqual(nil, err)
  t.AssertEqual(1, count)

  count = 0
  values = url.Values{}
  values.Set("comment", msg)
  t.POST(fmt.Sprintf(
    "/api/v0/posts/%d/comments", post.ID), values)
  err = db.Where("text = ?", msg).Find(
    &models.Comment{}).Count(&count).Error
  t.AssertEqual(nil, err)
  t.AssertEqual(1, count)

  count = 0
  values = url.Values{}
  t.POST(fmt.Sprintf(
    "/api/v0/posts/%d/likes/true", post.ID), values)
  err = db.Find(&models.Like{}).Count(&count).Error
  t.AssertEqual(nil, err)
  t.AssertEqual(1, count)

  aspect := struct{ID uint}{}
  d1 := &models.Person{}
  d2 := &models.Person{}
  values = url.Values{}
  values.Set("aspect_name", "test1")
  aspectBody := t.POST("/api/v0/aspects", values)
  err = json.Unmarshal(aspectBody, &aspect)
  t.AssertEqual(nil, err)
  err = db.Find(&models.Aspect{}).Count(&count).Error
  t.AssertEqual(nil, err)
  t.AssertEqual(1, count)

  if t.CI() {
    // test contact entity
    err = db.Where("author = ?", "d1@localhost:3000").Find(d1).Error
    t.AssertEqual(nil, err)
    err = db.Where("author = ?", "d2@localhost:3001").Find(d2).Error
    t.AssertEqual(nil, err)
    t.POST(fmt.Sprintf(
      "/api/v0/people/%d/aspects/%d",
      d1.ID, aspect.ID), url.Values{})
    err = db.Find(&models.AspectMembership{}).Count(&count).Error
    t.AssertEqual(nil, err)
    t.AssertEqual(1, count)
    t.POST(fmt.Sprintf(
      "/api/v0/people/%d/aspects/%d",
      d2.ID, aspect.ID), url.Values{})
    err = db.Find(&models.AspectMembership{}).Count(&count).Error
    t.AssertEqual(nil, err)
    t.AssertEqual(2, count)

    // wait some time to federate
    <-time.After(federation_timeout)

    for _, name := range ciDatabases {
      ciDB, err := t.DB(name)
      t.AssertEqual(nil, err)
      defer ciDB.Close()

      result := struct{Count int}{}
      tests := []struct{Scope *gorm.DB}{
        {Scope: ciDB.Raw("select count(*) as count from posts where text = ?", msg)},
        {Scope: ciDB.Raw("select count(*) as count from comments where text = ?", msg)},
        {Scope: ciDB.Raw("select count(*) as count from likes")},
        {Scope: ciDB.Raw("select count(*) as count from contacts")},
      }
      for i, test := range tests {
        err = test.Scope.Scan(&result).Error
        t.Assertf(err == nil, "#%d: expected nil, got '%+v'", i, err)
        t.Assertf(result.Count == 1, "#%d: expected 1, got %d", i, result.Count)
      }
    }
  }
}

func (t *FederationLocalTest) After() {
  db, err := models.OpenDatabase()
  t.AssertEqual(nil, err)
  defer db.Close()

  deleteModels := []interface{}{
    &models.Like{},
    &models.Comment{},
    &models.Post{},
    &models.AspectMembership{},
    &models.Aspect{},
  }

  for i, deleteModel := range deleteModels {
    err = db.Delete(deleteModel).Error
    t.Assertf(err == nil, "#%d: expected nil, got '%+v'", i, err)
  }
}
