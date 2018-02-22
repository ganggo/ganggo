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
  "encoding/xml"
  "github.com/revel/revel"
  run "github.com/revel/modules/jobs/app/jobs"
  "github.com/ganggo/ganggo/app/models"
  "github.com/ganggo/ganggo/app/helpers"
  federation "github.com/ganggo/federation"
)

func (dispatcher *Dispatcher) Contact(contact federation.EntityContact) {
  _, host, err := helpers.ParseAuthor(contact.Recipient)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  var person models.Person
  err = person.FindByAuthor(contact.Recipient)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  entityXml, err := xml.Marshal(contact)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  privKey, err := federation.ParseRSAPrivateKey(
    []byte(dispatcher.User.SerializedPrivateKey))
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  pubKey, err := federation.ParseRSAPublicKey(
    []byte(person.SerializedPublicKey))
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  payload, err := federation.EncryptedMagicEnvelope(
    privKey, pubKey, contact.Author, entityXml)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  run.Now(send{
    Host: host,
    Guid: person.Guid,
    Payload: payload,
  })
}
