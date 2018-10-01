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
  "git.feneas.org/ganggo/ganggo/app/models"
  fhelpers "git.feneas.org/ganggo/federation/helpers"
  federation "git.feneas.org/ganggo/federation"
)

func (dispatcher *Dispatcher) Contact(contact models.AspectMembership) {
  var person models.Person
  err := person.FindByID(contact.PersonID)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  endpoint := person.Inbox

  priv, err := fhelpers.ParseRSAPrivateKey(
    []byte(dispatcher.User.SerializedPrivateKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher Contact", err.Error(), err)
    return
  }

  pub, err := fhelpers.ParseRSAPublicKey(
    []byte(person.SerializedPublicKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher Contact", err.Error(), err)
    return
  }

  entity, err := federation.NewMessageContact(person.Pod.Protocol)
  if err != nil {
    revel.AppLog.Error("Dispatcher Contact", err.Error(), err)
    return
  }
  entity.SetAuthor(dispatcher.User.Person.Author)
  entity.SetRecipient(person.Author)
  entity.SetSharing(true)

  err = entity.Send(endpoint, priv, pub)
  if err != nil {
    revel.AppLog.Error("Dispatcher Contact", err.Error(), err)
    return
  }
}
