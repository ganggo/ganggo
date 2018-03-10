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
  "github.com/ganggo/ganggo/app/models"
  "github.com/ganggo/ganggo/app/helpers"
  federation "github.com/ganggo/federation"
  activitypub "github.com/ganggo/federation/activitypub"
)

type FetchAuthor struct {
  Author string
  Person models.Person
  Err error
}

func (fetch *FetchAuthor) Run() {
  guid := helpers.UuidFromSalt(fetch.Author)
  if fetch.AuthorLink() {
    err := fetch.Person.FindByGuid(guid)
    if err == nil {
      return
    }
  } else {
    err := fetch.Person.FindByAuthor(fetch.Author)
    if err == nil {
      return
    }
  }

  host, err := helpers.ParseHost(fetch.Author)
  if err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  // add host to pod list
  pod := models.Pod{Host: host}
  if err := pod.CreateOrFindHost(); err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  if fetch.AuthorLink() {
    fetch.activityPub(pod, guid)
  } else {
    fetch.diaspora(pod)
  }
}

func (fetch *FetchAuthor) AuthorLink() bool {
  return len(fetch.Author) > 3 && fetch.Author[:4] == "http"
}

func (fetch *FetchAuthor) activityPub(pod models.Pod, guid string) {
  var actor activitypub.ActivityActor
  err := federation.FetchJson("GET", fetch.Author, nil, &actor)
  if err != nil {
    revel.AppLog.Error(err.Error())
    fetch.Err = err
    return
  }

  author := actor.Id
  if actor.PreferredUsername != nil {
    author = *actor.PreferredUsername + "@" + pod.Host
  }

  if actor.PublicKey == nil {
    revel.AppLog.Error(err.Error())
    fetch.Err = errors.New("no public key available")
    return
  }

  fetch.Person = models.Person{
    Guid: guid,
    Author: author,
    SerializedPublicKey: actor.PublicKey.PublicKeyPem,
    PodID: pod.ID,
    Profile: models.Profile{
      Author: author,
      //Actor: actor.Id,
    },
  }

  if actor.Name != nil {
    fetch.Person.Profile.FirstName = *actor.Name
  }

  if actor.Icon != nil {
    fetch.Person.Profile.ImageUrl = actor.Icon.Url
  }

  if err = fetch.Person.Create(); err != nil {
    revel.AppLog.Error(err.Error())
    fetch.Err = err
    return
  }

  revel.AppLog.Debug("Added a new identity", "person", fetch.Person)
}

func (fetch *FetchAuthor) diaspora(pod models.Pod) {
  webFinger := federation.WebFinger{Host: pod.Host, Handle: fetch.Author}
  if err := webFinger.Discovery(); err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  var hcard federation.Hcard
  for _, link := range webFinger.Data.Links {
    if link.Rel == federation.WebFingerHcard {
      if err := hcard.Fetch(link.Href); err != nil {
        revel.AppLog.Error(err.Error())
        (*fetch).Err = err
        return
      }
    }
  }

  if hcard.Guid == "" || hcard.PublicKey == "" {
    (*fetch).Err = errors.New("Something went wrong! Hcard struct is empty.")
    return
  }

  (*fetch).Person = models.Person{
    Guid: hcard.Guid,
    Author: fetch.Author,
    SerializedPublicKey: hcard.PublicKey,
    PodID: pod.ID,
    Profile: models.Profile{
      Author: fetch.Author,
      FullName: hcard.FullName,
      Searchable: hcard.Searchable,
      FirstName: hcard.FirstName,
      LastName: hcard.LastName,
      ImageUrl: hcard.Photo,
    },
    Contacts: models.Contacts{},
  }

  if err := fetch.Person.Create(); err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  revel.AppLog.Debug("Added a new identity", "person", fetch.Person)
}
