package notifier
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
  "net/url"
)

type Telegram struct {
  ID string
  Text string
}

type TelegramErrorResponse struct {
  Ok bool `json:"ok"`
  ErrorCode int `json:"error_code"`
  Description string `json:"description"`
}

func (telegram Telegram) Send() {
  token, ok := revel.Config.String("telegram.token"); if !ok {
    revel.AppLog.Error("Telegram", "Missing telegram settings!")
    return
  }

  endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
  data := url.Values{}
  data.Set("chat_id", telegram.ID)
  data.Set("text", telegram.Text)
  data.Set("parse_mode", "html")

  resp, err := http.PostForm(endpoint, data)
  if err != nil {
    revel.AppLog.Error("Telegram", err.Error(), err)
    return
  }

  if resp.StatusCode != http.StatusOK {
    var telegramResp TelegramErrorResponse
    err := json.NewDecoder(resp.Body).Decode(&telegramResp)
    if err == nil && !telegramResp.Ok {
      revel.AppLog.Error("Telegram", "telegramResp", telegramResp)
      return
    }
  }
  resp.Body.Close()
}
