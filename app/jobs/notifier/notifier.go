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

import "github.com/revel/revel"

type Notifier struct {
  Messages []interface{}
}

func (notifier Notifier) Run() {
  revel.Config.SetSection("ganggo")

  for _, message := range notifier.Messages {
    switch job := message.(type) {
    case Mail:
      if revel.Config.BoolDefault("mail.enabled", false) {
        job.Send()
      }
    case Telegram:
      if revel.Config.BoolDefault("telegram.enabled", false) {
        job.Send()
      }
    }
  }
}
