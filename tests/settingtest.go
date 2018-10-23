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
  "git.feneas.org/ganggo/ganggo/app/routes"
  "git.feneas.org/ganggo/ganggo/app/models"
)

type SettingTest struct {
  GnggTestSuite
}

func (t *SettingTest) Before() {
  t.ClearDB()
  t.CreateUser()
}

var settingTests = []struct{
  Values map[string]string
  Settings map[models.UserSettingKey]string
}{
  {
    Values: map[string]string{"lang": "en"},
    Settings: map[models.UserSettingKey]string{
      models.UserSettingLanguage: "en",
    },
  },
  {
    Values: map[string]string{"email": "test@local.host"},
    Settings: map[models.UserSettingKey]string{
      models.UserSettingMailAddress: "test@local.host",
      models.UserSettingMailVerified: "",
    },
  },
}

func (t *SettingTest) TestSettings() {
  db, err := models.OpenDatabase()
  t.Assertf(err == nil, "Expected nil, got '%+v'", err)
  defer db.Close()

  id := t.UserID(t.AccessToken())

  for i, test := range settingTests {
    values := url.Values{}
    for key, value := range test.Values {
      values.Set(key, value)
    }
    t.POST(routes.Setting.Update(), values)

    var ii int
    for key, value := range test.Settings {
      var setting models.UserSetting
      err = db.Where("setting_key = ? and user_id = ?", key, id).
        First(&setting).Error

      var statement bool
      switch key {
      case models.UserSettingMailVerified:
        statement = len(setting.Value) == 64
      default:
        statement = setting.Value == value
      }

      t.Assertf(err == nil, "#%d.%d: Expected nil, got '%+v' %d %d", i, ii, err)
      t.Assertf(statement, "#%d.%d: Expected '%s', got '%s'",
        i, ii, value, setting.Value)

      ii += 1
    }
  }
}
