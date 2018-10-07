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
  "git.feneas.org/ganggo/federation"
)

type Dispatcher struct {
  User models.User
  Message interface{}
  Retract bool
}

func (dispatcher Dispatcher) Run() {
  switch entity := dispatcher.Message.(type) {
  case models.AspectMembership:
    revel.AppLog.Debug("Starting contact dispatcher")
    dispatcher.Contact(entity)
  case models.Post:
    revel.AppLog.Debug("Starting post dispatcher")
    dispatcher.StatusMessage(entity)
  case models.Comment:
    revel.AppLog.Debug("Starting comment dispatcher")
    dispatcher.Comment(entity)
  case models.Like:
    revel.AppLog.Debug("Starting like dispatcher")
    dispatcher.Like(entity)
  // relaying entities
  case federation.MessageComment:
    revel.AppLog.Debug("Starting relay comment dispatcher")
    dispatcher.RelayComment(entity)
  case federation.MessageLike:
    revel.AppLog.Debug("Starting relay like dispatcher")
    dispatcher.RelayLike(entity)
  case federation.MessageRetract:
    revel.AppLog.Debug("Starting relay retraction dispatcher")
    dispatcher.RelayRetraction(entity)
  default:
    revel.AppLog.Error("Unknown entity type in dispatcher!")
  }
}

// findRecipients will return a list of users we should relay a post to
func (dispatcher *Dispatcher) findRecipients(parentPost *models.Post, parentUser *models.User) ([]models.Person, error) {
  if parentPost != nil && parentUser != nil {
    if parentPost.Public {
      // everyone we are sharing with
      return dispatcher.findPublicEndpoints()
    } else {
      // it is local we know all recipients
      // lets relay the message to all remote servers
      var visibility models.AspectVisibility
      err := visibility.FindByPost(*parentPost)
      if err != nil {
        revel.AppLog.Error("Dispatcher findRecipients", err.Error(), err)
        return []models.Person{}, err
      }

      var aspect models.Aspect
      err = aspect.FindByID(visibility.AspectID)
      if err != nil {
        revel.AppLog.Error("Dispatcher findRecipients", err.Error(), err)
        return []models.Person{}, err
      }
      var persons []models.Person
      for _, member := range aspect.Memberships {
        var person models.Person
        err = person.FindByID(member.PersonID)
        if err != nil {
          revel.AppLog.Error("Dispatcher findRecipients", err.Error(), err)
          continue
        }
        persons = append(persons, person)
      }
      return persons, nil
    }
  } else if parentPost != nil {
    // it is not local just send it to
    // the remote server it should handle the rest
    var persons = []models.Person{parentPost.Person}
    if !parentPost.Public {
      // in case of AP we will fetch known visibilties as well
      var visibilities models.Visibilities
      err := visibilities.FindByPost(*parentPost)
      if err == nil {
        for _, visibility := range visibilities {
          persons = append(persons, visibility.Person)
        }
      } else {
        revel.AppLog.Error("Dispatcher findRecipients", err.Error(), err)
      }
    }
    return persons, nil
  }
  return []models.Person{}, nil
}

// findPublicEndpoints will fetch all remote endpoints known to the server
// NOTE this can become a pretty heavy job needs some brain-storming
func (dispatcher *Dispatcher) findPublicEndpoints() (persons []models.Person, err error) {
  var pods models.Pods
  err = pods.FindAll()
  if err != nil {
    return
  }

  for _, pod := range pods {
    var person models.Person
    err = person.FindFirstByPodID(pod.ID)
    if err != nil {
      continue
    }
    persons = append(persons, person)
  }
  return
}
