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
  "fmt"
  "git.feneas.org/ganggo/ganggo/app/jobs"
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/gorm"
  "time"
)

type JobsTest struct {
  GnggTestSuite
}

var tgReceiverTests = []struct{
  Text string
  Expected string
}{
  {
    Text: "hi",
    Expected: "",
  },
  {
    Text: "/start",
    Expected: "",
  },
  {
    Text: "/stop",
    Expected: "",
  },
}

var sessionTests = []struct {
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
  for i, test := range sessionTests {
    err = db.Create(&models.Session{
      CreatedAt: test.CreatedAt,
      Token: test.Token,
      UserID: t.UserID(token),
    }).Error
    t.Assertf(err == nil, "#%d. session creation failed: %+v", i, err)
  }

  sessionJob := jobs.Session{}
  sessionJob.Run()

  for i, test := range sessionTests {
    err = db.Where("token = ?", test.Token).First(&models.Session{}).Error
    t.Assertf(err == test.ExpectedErr,
      "#%d: expected '%+v', got '%+v'", i, test.ExpectedErr, err)
  }
}

func (t *JobsTest) TestTelegramReceiver() {
  var tmpl = `{"message":{"from":{"id":1234},"chat":{"id":4321},"text":"%s"}}`

  db, err := models.OpenDatabase()
  t.Assertf(err == nil, "Expected nil, got '%+v'", err)
  defer db.Close()

  var setting models.UserSetting
  id := t.UserID(t.AccessToken())
  err = db.Where("user_id = ? and setting_key = ?", id,
    models.UserSettingTelegramVerified).First(&setting).Error
  t.Assertf(err == nil, "Expected nil, got '%+v'", err)

  tgReceiverTests = append(tgReceiverTests, struct{Text, Expected string}{
    Text: setting.Value,
    Expected: "4321",
  })

  for i, test := range tgReceiverTests {
    (jobs.TelegramReceiver{
      Body: []byte(fmt.Sprintf(tmpl, test.Text))}).Run()

    var setting models.UserSetting
    db.Where("setting_key = ? and value = ?",
      models.UserSettingTelegramID, "4321").Find(&setting)

    if test.Expected == "" {
      t.Assertf(setting.UserID == 0,
        "#%d: expected no record, got '%+v'", i, setting)
    } else {
      t.Assertf(setting.UserID != 0,
        "#%d: expected '%s', got no record %+v", i, test.Expected, test)
    }
  }

  t.Assertf(len(tgReceiverTests) == 4,
    "Expected four entries, got %d", len(tgReceiverTests))
}
