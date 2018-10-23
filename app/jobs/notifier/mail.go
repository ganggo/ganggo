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
  "crypto/tls"
  "github.com/go-gomail/gomail"
  "bytes"
)

type Mail struct {
  To, Subject, Body, Lang string
}

func (mail Mail) Send() {
  msg := gomail.NewMessage()
  revel.Config.SetSection("ganggo")
  host, ok := revel.Config.String("mail.host"); if !ok {
    revel.AppLog.Error("Missing mail settings!")
    return
  }
  port, ok := revel.Config.Int("mail.port"); if !ok {
    revel.AppLog.Error("Missing mail settings!")
    return
  }
  insecure := revel.Config.BoolDefault("mail.insecure", false)
  username, ok := revel.Config.String("mail.username"); if !ok {
    revel.AppLog.Error("Missing mail settings!")
    return
  }
  password, ok := revel.Config.String("mail.password"); if !ok {
    revel.AppLog.Error("Missing mail settings!")
    return
  }
  from, ok := revel.Config.String("mail.from"); if !ok {
    revel.AppLog.Error("Missing mail settings!")
    return
  }

  tmpl, err := revel.MainTemplateLoader.TemplateLang(
    "notifier/mail.html", mail.Lang)
  if err != nil {
    revel.AppLog.Error("Mail", err.Error(), err)
    return
  }

  var buf bytes.Buffer
  err = tmpl.Render(&buf, map[string]interface{}{
    revel.CurrentLocaleViewArg: mail.Lang,
    "Text": mail.Body})
  if err != nil {
    revel.AppLog.Error("Mail", err.Error(), err)
    return
  }

  msg.SetHeader("From", from)
  msg.SetHeader("To", mail.To)
  msg.SetHeader("Subject", mail.Subject)
  msg.SetBody("text/html", buf.String())

  dialer := gomail.NewDialer(host, port, username, password)
  if insecure {
    dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
  }

  err = dialer.DialAndSend(msg)
  if err != nil {
    revel.AppLog.Error("Mail", err.Error(), err)
    return
  }
}


