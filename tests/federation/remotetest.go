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
  "time"
  "os/exec"
  "gopkg.in/ganggo/ganggo.v0/app/models"
)

type FederationRemoteTest struct {
  FederationSuite
}

func (t *FederationRemoteTest) Before() {
  println("Setup user relations and access token")
  t.AssertEqual(nil, t.SetupUserRelations())
}

func (t *FederationRemoteTest) TestRemote() {
  if !t.CI() {
    // skip non-CI
    return
  }

  for _, name := range ciDatabases {
    // share with ganggo user
    cmd := exec.Command("docker", "exec", name,
      "bundle", "exec", "rails", "runner", "share_with.rb", name, handle)
    err := cmd.Run()
    t.AssertEqual(nil, err)
    // send post, like and comments
    cmd = exec.Command("docker", "exec", name,
      "bundle", "exec", "rails", "runner", "post_like_comment.rb", name)
    err = cmd.Run()
    t.AssertEqual(nil, err)
  }
  // wait some time to federate
  <-time.After(5 * time.Second)

  db, err := models.OpenDatabase()
  t.AssertEqual(nil, err)
  defer db.Close()

  var count int
  err = db.Find(&models.Posts{}).Count(&count).Error
  t.AssertEqual(nil, err)
  t.AssertEqual(4, count)

  count = 0
  err = db.Find(&models.Likes{}).Count(&count).Error
  t.AssertEqual(nil, err)
  t.AssertEqual(4, count)

  count = 0
  err = db.Find(&models.Comments{}).Count(&count).Error
  t.AssertEqual(nil, err)
  t.AssertEqual(4, count)
}

func (t *FederationRemoteTest) After() {
  if t.CI() {
    db, err := models.OpenDatabase()
    t.AssertEqual(nil, err)
    defer db.Close()

    err = db.Delete(&models.Comment{}).Error
    t.AssertEqual(nil, err)
    err = db.Delete(&models.Like{}).Error
    t.AssertEqual(nil, err)
    err = db.Delete(&models.Post{}).Error
    t.AssertEqual(nil, err)
  }
}
