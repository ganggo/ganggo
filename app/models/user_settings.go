package models
//
// GangGo Application Server
// Copyright (C) 2018 Lukas Matt <lukas@zauberstuhl.de>
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

import "github.com/revel/revel"

type UserSettingKey int

const (
  UserSettingLanguage UserSettingKey = iota
  UserSettingMailVerified UserSettingKey = iota + 10
  UserSettingMailAddress
  UserSettingTelegramVerified UserSettingKey = iota + 20
  UserSettingTelegramID
)

type UserSetting struct {
  ID uint `gorm:"primary_key"`
  UserID uint

  // NOTE Key is a special word in MySQL
  Key UserSettingKey `gorm:"column:setting_key"`
  Value string
}

type UserSettings []UserSetting

func (settings UserSettings) GetValue(key UserSettingKey) string {
  for _, setting := range settings {
    if setting.Key == key {
      return setting.Value
    }
  }
  return ""
}

func (settings *UserSettings) Update() (err []error) {
  for _, setting := range *settings {
    err = append(err, setting.Update())
  }
  return err
}

func (setting *UserSetting) FindByKeyValue(key UserSettingKey, value string) error {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return err
  }
  defer db.Close()

  return db.Where("setting_key = ? and value = ?",
    key, value).First(setting).Error
}

func (setting *UserSetting) Update() error {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return err
  }
  defer db.Close()

  var dbSetting UserSetting
  err = db.Where("user_id = ? and setting_key = ?",
    setting.UserID, setting.Key).First(&dbSetting).Error
  if err == nil {
    setting.ID = dbSetting.ID
    return db.Save(setting).Error
  }
  return db.Create(setting).Error
}

func (settings *UserSettings) Delete() (err []error) {
  for _, setting := range *settings {
    err = append(err, setting.Delete())
  }
  return err
}

func (setting *UserSetting) Delete() error {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return err
  }
  defer db.Close()

  return db.Where("user_id = ? and setting_key = ?",
    setting.UserID, setting.Key).Delete(UserSetting{}).Error
}
