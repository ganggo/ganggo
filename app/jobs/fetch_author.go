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
// along with this program.  If not, see <http://www.gnf.org/licenses/>.
//

import (
  "errors"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/ganggo/app/helpers"
  federation "git.feneas.org/ganggo/federation"
  "fmt"
)

type FetchAuthor struct {
  Author string
  Person models.Person
  Err error
}

func (fetch *FetchAuthor) Run() {
  err := fetch.Person.FindByAuthor(fetch.Author)
  if err == nil {
    return
  }

  host, err := helpers.ParseHost(fetch.Author)
  if err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  var actorId string
  var hcard federation.Hcard
  var proto federation.Protocol
  http := len(fetch.Author) > 3 && fetch.Author[:4] == "http"
  if http {
    proto = federation.ActivityPubProtocol
    actorId = fetch.Author
  } else {
    webFinger := federation.WebFinger{Host: host, Handle: fetch.Author}
    if err := webFinger.Discovery(); err != nil {
      revel.AppLog.Error(err.Error())
      (*fetch).Err = err
    }
    proto = webFinger.Protocol()

    for _, link := range webFinger.Data.Links {
      if proto == federation.DiasporaProtocol &&
      link.Rel == federation.WebFingerHcard {
        if err := hcard.Fetch(link.Href); err != nil {
          revel.AppLog.Error(err.Error())
          (*fetch).Err = err
          return
        }
        break
      }
      if proto == federation.ActivityPubProtocol &&
      link.Rel == federation.WebFingerSelf {
        actorId = link.Href
        break
      }
    }
  }

  err = fetch.Person.FindByGuid(helpers.UuidFromSalt(actorId))
  if err == nil {
    return
  }

  switch proto {
  case federation.DiasporaProtocol:
    endpoint := fmt.Sprintf("https://%s/receive", host)

    if hcard.Guid == "" || hcard.PublicKey == "" {
      (*fetch).Err = errors.New("Something went wrong! Hcard struct is empty.")
      return
    }

    pod := models.Pod{Host: host, Protocol: proto,
      Inbox: fmt.Sprintf("%s/public", endpoint)}
    if err := pod.CreateOrFindHost(); err != nil {
      revel.AppLog.Error(err.Error())
      (*fetch).Err = err
      return
    }

    imageUrl := "/public/img/avatar.png"
    if hcard.Photo != "" {
      imageUrl = hcard.Photo
    }

    (*fetch).Person = models.Person{
      Guid: hcard.Guid,
      Author: fetch.Author,
      Inbox: fmt.Sprintf("%s/users/%s", endpoint, hcard.Guid),
      SerializedPublicKey: hcard.PublicKey,
      PodID: pod.ID,
      Profile: models.Profile{
        Author: fetch.Author,
        FullName: hcard.FullName,
        Searchable: hcard.Searchable,
        FirstName: hcard.FirstName,
        LastName: hcard.LastName,
        ImageUrl: imageUrl,
      },
      Contacts: models.Contacts{},
    }

    if err := fetch.Person.Create(); err != nil {
      revel.AppLog.Error(err.Error())
      (*fetch).Err = err
      return
    }

    revel.AppLog.Debug("Added a new identity", "person", fetch.Person)
  case federation.ActivityPubProtocol:
    var actor federation.ActivityPubActor
    err := federation.FetchJson("GET", actorId, nil, &actor)
    if err != nil {
      revel.AppLog.Error(err.Error())
      (*fetch).Err = err
      return
    }

    if actor.PublicKey == nil {
      (*fetch).Err = errors.New("PublicKey cannot be empty!")
      revel.AppLog.Error(fetch.Err.Error())
      return
    }

    var sharedInbox string
    if actor.Endpoints != nil {
      sharedInbox = actor.Endpoints.SharedInbox
    }

    pod := models.Pod{Host: host, Protocol: proto, Inbox: sharedInbox}
    if err := pod.CreateOrFindHost(); err != nil {
      revel.AppLog.Error(err.Error())
      (*fetch).Err = err
      return
    }

    fetch.Person = models.Person{
      Guid: helpers.UuidFromSalt(actor.Id),
      Author: actor.Id,
      SerializedPublicKey: actor.PublicKey.PublicKeyPem,
      PodID: pod.ID,
      Inbox: actor.Inbox,
      Profile: models.Profile{
        Author: actor.Id,
      },
    }

    if actor.PreferredUsername != nil {
      fetch.Person.Profile.FirstName = *actor.PreferredUsername
    } else if actor.Name != nil {
      fetch.Person.Profile.FirstName = *actor.Name
    }

    if actor.Summary != nil {
      fetch.Person.Profile.Bio = *actor.Summary
    }

    if actor.Icon != nil {
      fetch.Person.Profile.ImageUrl = actor.Icon.Url
    } else {
      fetch.Person.Profile.ImageUrl = "/public/img/avatar.png"
    }

    if err = fetch.Person.Create(); err != nil {
      revel.AppLog.Error(err.Error())
      (*fetch).Err = err
      return
    }

    revel.AppLog.Debug("Added a new identity", "person", fetch.Person)
  default:
    (*fetch).Err = errors.New("Something went wrong! Unknown protocol...")
  }
}
