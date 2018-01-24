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
  "gopkg.in/ganggo/ganggo.v0/app/models"
  federation "gopkg.in/ganggo/federation.v0"
)

type Receiver struct {
  Message federation.Message
  Entity federation.Entity
  Guid string
}

func (receiver *Receiver) Run() {
  // search for sender and check his signature
  person, ok := receiver.CheckAuthor(receiver.Message.Sig.KeyId)
  if !ok || !valid(person, receiver.Message, "") {
    return
  }

  switch entity := receiver.Entity.Data.(type) {
  case federation.EntityRetraction:
    if _, ok := receiver.CheckAuthor(entity.Author); ok {
      revel.AppLog.Debug("Starting retraction receiver")
      receiver.Retraction(entity)
    }
  case federation.EntityProfile:
    if _, ok := receiver.CheckAuthor(entity.Author); ok {
      revel.AppLog.Debug("Starting profile receiver")
      receiver.Profile(entity)
    }
  case federation.EntityReshare:
    if _, ok := receiver.CheckAuthor(entity.Author); ok {
      revel.AppLog.Debug("Starting reshare receiver")
      receiver.Reshare(entity)
    }
  case federation.EntityStatusMessage:
    if _, ok := receiver.CheckAuthor(entity.Author); ok {
      revel.AppLog.Debug("Starting status message receiver")
      receiver.StatusMessage(entity)
    }
  case federation.EntityComment:
    if person, ok := receiver.CheckAuthor(entity.Author); ok {
      revel.AppLog.Debug("Starting comment receiver")
      // validate author_signature
      if valid(person, entity, receiver.Entity.SignatureOrder) {
        receiver.Comment(entity)
      } else {
        revel.AppLog.Error("invalid sig", "entity", entity)
      }
    }
  case federation.EntityLike:
    if person, ok := receiver.CheckAuthor(entity.Author); ok {
      revel.AppLog.Debug("Starting like receiver")
      // validate author_signature
      if valid(person, entity, receiver.Entity.SignatureOrder) {
        receiver.Like(entity)
      }
    }
  case federation.EntityContact:
    if _, ok := receiver.CheckAuthor(entity.Author); ok {
      revel.AppLog.Debug("Starting contact receiver")
      receiver.Contact(entity)
    }
  default:
    revel.AppLog.Error("No matching entity found", "entity", receiver.Entity)
  }
}

func (receiver *Receiver) CheckAuthor(author string) (models.Person, bool) {
  // Will try fetching author from remote
  // if he doesn't exist locally
  fetch := FetchAuthor{Author: author}; fetch.Run()
  if fetch.Err != nil {
    revel.AppLog.Error("Cannot fetch author", "error", fetch.Err)
  }
  return fetch.Person, fetch.Err == nil
}

func valid(person models.Person, entity federation.SignatureInterface, order string) bool {
  pubKey, err := federation.ParseRSAPublicKey(
    []byte(person.SerializedPublicKey))
  if err != nil {
    revel.AppLog.Error(err.Error())
    return false
  }

  // verify sender signature
  var signature federation.Signature
  if !signature.New(entity).Verify(order, pubKey) {
    revel.AppLog.Warn("Signature verification failed", "err", signature.Err)
    return false
  }
  revel.AppLog.Debug("Valid signature", "guid", person.Guid)
  return true
}
