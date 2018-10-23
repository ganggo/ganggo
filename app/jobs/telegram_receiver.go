package jobs
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

import (
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "encoding/json"
  "strconv"
)

type TelegramReceiver struct {
  Body []byte
}

type TelegramMessage struct {
  Message struct {
    From struct {
      ID int `json:"id"`
    } `json:"from"`
    Chat struct {
      ID int `json:"id"`
    } `json:"chat"`
    Text string `json:"text"`
  } `json:"message"`
}

func (receiver TelegramReceiver) Run() {
  telegramMsg := TelegramMessage{}
  err := json.Unmarshal(receiver.Body, &telegramMsg)
  if err !=  nil {
    revel.AppLog.Error("TelegramReceiver", "error", err)
    return
  }

  var tgSetting models.UserSetting
  if telegramMsg.Message.Text == "/stop" {
    err = tgSetting.FindByKeyValue(models.UserSettingTelegramID,
      strconv.Itoa(telegramMsg.Message.Chat.ID))
    if err ==  nil {
      token, err := helpers.Token()
      if err != nil {
        revel.AppLog.Error("TelegramReceiver", "error", err)
        return
      }
      tgSetting.Key = models.UserSettingTelegramVerified
      tgSetting.Value = token
      err = tgSetting.Update()
      if err != nil {
        revel.AppLog.Error("TelegramReceiver", "error", err)
        return
      }
      tgSetting.Key = models.UserSettingTelegramID
      err = tgSetting.Delete()
      if err != nil {
        revel.AppLog.Error("TelegramReceiver", "error", err)
        return
      }
    }
  } else if len(telegramMsg.Message.Text) == 64 {
    err = tgSetting.FindByKeyValue(
      models.UserSettingTelegramID, strconv.Itoa(telegramMsg.Message.Chat.ID))
    // ignore it if someone already uses the account
    if err != nil {
      err = tgSetting.FindByKeyValue(
        models.UserSettingTelegramVerified, telegramMsg.Message.Text)
      if err ==  nil {
        var tgID models.UserSetting
        tgID.UserID = tgSetting.UserID
        tgID.Key = models.UserSettingTelegramID
        tgID.Value = strconv.Itoa(telegramMsg.Message.Chat.ID)
        tgSetting.Value = "true"

        settings := models.UserSettings{tgSetting, tgID}
        errList := settings.Update()
        for _, err := range errList {
          if err !=  nil {
            revel.AppLog.Error("TelegramReceiver", "error", err)
            return
          }
        }
      }
    }
  }
}
