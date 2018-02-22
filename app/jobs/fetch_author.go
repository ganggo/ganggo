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
)

type FetchAuthor struct {
  Author string
  Person models.Person
  Err error
}

func (fetch *FetchAuthor) Run() {
  var person models.Person
  _, host, err := helpers.ParseAuthor(fetch.Author)
  if err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  if err = person.FindByAuthor(fetch.Author); err == nil {
    fetch.Person = person
    return
  }

  // add host to pod list
  pod := models.Pod{Host: host}
  if err := pod.CreateOrFindHost(); err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  webFinger := federation.WebFinger{Host: host, Handle: fetch.Author}
  if err = webFinger.Discovery(); err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  var hcard federation.Hcard
  for _, link := range webFinger.Data.Links {
    if link.Rel == federation.WebFingerHcard {
      if err = hcard.Fetch(link.Href); err != nil {
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
      ImageUrlSmall: hcard.PhotoSmall,
      ImageUrlMedium: hcard.PhotoMedium,
    },
    Contacts: models.Contacts{},
  }

  db, err := models.OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }
  defer db.Close()

  if err = db.Create(&fetch.Person).Error; err != nil {
    revel.AppLog.Error(err.Error())
    (*fetch).Err = err
    return
  }

  revel.AppLog.Debug("Added a new identity", "person", fetch.Person)
}
