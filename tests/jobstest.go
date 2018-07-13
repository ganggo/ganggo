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
  "git.feneas.org/ganggo/ganggo/app/jobs"
  "git.feneas.org/ganggo/ganggo/app/models"
  "gopkg.in/ganggo/gorm.v2"
  "time"
)

type JobsTest struct {
  GnggTestSuite
}

var tests = []struct {
  Token string
  CreatedAt time.Time
  ExpectedErr error
}{
  {
    Token: "4321no1",
    CreatedAt: time.Now().AddDate(0, 0, -3),
    ExpectedErr: gorm.ErrRecordNotFound,
  },
  {
    Token: "4321no2",
    CreatedAt: time.Now().AddDate(0, 0, -1),
    ExpectedErr: nil,
  },
  {
    Token: "4321no3",
    CreatedAt: time.Now(),
    ExpectedErr: nil,
  },
}

func (t *JobsTest) Before() {
  t.ClearDB()
  t.CreateUser()
}

func (t *JobsTest) TestSession() {
  db, err := models.OpenDatabase()
  t.Assertf(err == nil, "Expected nil, got '%+v'", err)
  defer db.Close()

  token := t.AccessToken()
  for i, test := range tests {
    err = db.Create(&models.Session{
      CreatedAt: test.CreatedAt,
      Token: test.Token,
      UserID: t.UserID(token),
    }).Error
    t.Assertf(err == nil, "#%d. session creation failed: %+v", i, err)
  }

  sessionJob := jobs.Session{}
  sessionJob.Run()

  for i, test := range tests {
    err = db.Where("token = ?", test.Token).First(&models.Session{}).Error
    t.Assertf(err == test.ExpectedErr,
      "#%d: expected '%+v', got '%+v'", i, test.ExpectedErr, err)
  }
}
