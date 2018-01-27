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

  // wait some time to federate
  <-time.After(5 * time.Second)
  if t.CI() {
    for _, name := range ciDatabases {
      ciDB, err := t.DB(name)
      t.AssertEqual(nil, err)
      defer ciDB.Close()

      result := struct {Count int}{}
      err = ciDB.Raw(
        "select count(*) as count from posts where text = ?", msg,
      ).Scan(&result).Error
      t.AssertEqual(nil, err)
      t.AssertEqual(1, result.Count)

      err = ciDB.Raw(
        "select count(*) as count from comments where text = ?", msg,
      ).Scan(&result).Error
      t.AssertEqual(nil, err)
      t.AssertEqual(1, result.Count)

      err = ciDB.Raw(
        "select count(*) as count from likes").Scan(&result).Error
      t.AssertEqual(nil, err)
      t.AssertEqual(1, result.Count)
    }
  }
}

func (t *FederationLocalTest) After() {
  db, err := models.OpenDatabase()
  t.AssertEqual(nil, err)
  defer db.Close()

  err = db.Delete(&models.Like{}).Error
  t.AssertEqual(nil, err)

  err = db.Delete(&models.Comment{}).Error
  t.AssertEqual(nil, err)

  err = db.Delete(&models.Post{}).Error
  t.AssertEqual(nil, err)
}
