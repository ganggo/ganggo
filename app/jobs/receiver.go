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
  federation "gopkg.in/ganggo/federation.v0"
)

type Receiver struct {
  Entity federation.Entity
  Guid string
}

func (receiver *Receiver) Run() {
  // TODO signature verification for entities missing !!!
  switch entity := receiver.Entity.Data.(type) {
  case federation.EntityRetraction:
    revel.AppLog.Debug("Starting retraction receiver")
    receiver.Retraction(entity)
  case federation.EntityProfile:
    revel.AppLog.Debug("Starting profile receiver")
    receiver.Profile(entity)
  case federation.EntityReshare:
    revel.AppLog.Debug("Starting reshare receiver")
    receiver.Reshare(entity)
  case federation.EntityStatusMessage:
    revel.AppLog.Debug("Starting status message receiver")
    receiver.StatusMessage(entity)
  case federation.EntityComment:
    revel.AppLog.Debug("Starting comment receiver")
    receiver.Comment(entity)
  case federation.EntityLike:
    revel.AppLog.Debug("Starting like receiver")
    receiver.Like(entity)
  case federation.EntityContact:
    revel.AppLog.Debug("Starting contact receiver")
    receiver.Contact(entity)
  default:
    revel.AppLog.Error("No matching entity found", "entity", receiver.Entity)
  }
}
