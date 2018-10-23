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
  "encoding/json"
  "fmt"
  "net/http"
)

type TelegramWebhook struct {
  Token, Url string
}

type TelegramErrorResponse struct {
  Ok bool `json:"ok"`
  ErrorCode int `json:"error_code"`
  Description string `json:"description"`
}

func (webhook TelegramWebhook) Run() {
  endpoint := fmt.Sprintf(
    "https://api.telegram.org/bot%s/setWebhook?url=%s/receive/telegram/%s&allowed_updates=message",
    webhook.Token, webhook.Url, webhook.Token)
  resp, err := http.Get(endpoint)
  if err != nil {
    revel.AppLog.Error("TelegramWebhook", err.Error(), err)
    return
  }

  if resp.StatusCode != http.StatusOK {
    var telegramResp TelegramErrorResponse
    err := json.NewDecoder(resp.Body).Decode(&telegramResp)
    if err == nil && !telegramResp.Ok {
      revel.AppLog.Error("TelegramWebhook", "resp", telegramResp)
      return
    }
  }
  resp.Body.Close()
}
