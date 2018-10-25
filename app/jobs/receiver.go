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
  federation "git.feneas.org/ganggo/federation"
  "git.feneas.org/ganggo/federation/helpers"
)

type Receiver struct {
  Message federation.Message
  Guid string
}

func (receiver Receiver) Run() {
  base := receiver.Message.Entity()
  if base == nil {
    revel.AppLog.Warn("Entity is nil!", "message", receiver.Message)
    return
  }

  // Check and search for author in the database
  // if it doesn't exists lookup the network
  person, ok := receiver.CheckAuthor(base.Author()); if !ok { return }

  pub, err := helpers.ParseRSAPublicKey(
    []byte(person.SerializedPublicKey))
  if err != nil {
    revel.AppLog.Error("Dispatcher", err.Error(), err)
    return
  }
  // Check signature and abort if it fails
  if !receiver.Message.ValidSignature(pub) {
    revel.AppLog.Debug("Invalid signature", "entity", base)
    return
  }

  // if the pod was deactivated and we receive a
  // new valid message we can set him active again
  if person.Pod.ID > 0 && !person.Pod.Alive {
    person.Pod.Alive = true
    err = person.Pod.Save()
    if err != nil {
      revel.AppLog.Error("Dispatcher", "error", err, "pod", person.Pod)
    }
  }

  switch entity := base.(type) {
  case federation.MessageContact:
    revel.AppLog.Debug("Starting contact receiver")
    receiver.Contact(entity)
  case federation.MessageRetract:
    revel.AppLog.Debug("Starting retraction receiver")
    receiver.Retraction(entity)
  case federation.MessagePost:
    revel.AppLog.Debug("Starting status message receiver")
    receiver.StatusMessage(entity)
  case federation.MessageReshare:
    revel.AppLog.Debug("Starting reshare receiver")
    receiver.Reshare(entity)
  case federation.MessageComment:
    revel.AppLog.Debug("Starting comment receiver")
    receiver.Comment(entity)
  case federation.MessageLike:
    revel.AppLog.Debug("Starting like receiver")
    receiver.Like(entity)
  case federation.MessageProfile:
    revel.AppLog.Debug("Starting profile receiver")
    receiver.Profile(entity)
  default:
    revel.AppLog.Error("No matching entity found", "msg", receiver.Message)
  }
}

func (receiver *Receiver) CheckAuthor(author string) (models.Person, bool) {
  // Will try fetching author from remote
  // if he doesn't exist locally
  fetch := FetchAuthor{Author: author}; fetch.Run()
  if fetch.Err != nil {
    revel.AppLog.Debug("Cannot fetch author", "error", fetch.Err)
  }
  return fetch.Person, fetch.Err == nil
}
