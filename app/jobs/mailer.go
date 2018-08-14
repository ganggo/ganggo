package jobs
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
  "github.com/revel/revel"
  //"gopkg.in/ganggo/ganggo.v0/app/models"
  //federation "gopkg.in/ganggo/federation.v0"
  "net/smtp"
  "strings"
)

type Mailer struct {
  Recipients []string
  Body []byte
}

func (mailer Mailer) Run() {
  revel.Config.SetSection("ganggo")
  username, found := revel.Config.String("mail.username"); if !found {
    revel.AppLog.Debug("Cannot find config value!", "mail.username", found)
    return
  }
  password, found := revel.Config.String("mail.password"); if !found {
    revel.AppLog.Debug("Cannot find config value!", "mail.password", found)
    return
  }
  host, found := revel.Config.String("mail.host"); if !found {
    revel.AppLog.Debug("Cannot find config value!", "mail.host", found)
    return
  }
  sender, found := revel.Config.String("mail.sender"); if !found {
    revel.AppLog.Debug("Cannot find config value!", "mail.sender", found)
    return
  }

  conInfo := strings.Split(host, ":")
  if len(conInfo) != 2 {
    revel.AppLog.Debug("Hostname should be in the format example.org:25!")
    return
  }

  auth := smtp.PlainAuth("", username, password, conInfo[0])
  if err := smtp.SendMail(
    host, auth, sender,
    mailer.Recipients, mailer.Body,
  ); err != nil {
    revel.AppLog.Error("Sending mail failed",
      "recipients", mailer.Recipients, "body", string(mailer.Body))
  }
}
