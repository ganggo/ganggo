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
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  federation "gopkg.in/ganggo/federation.v0"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type FetchAuthor struct {
  Author string
  Person *models.Person
  Err error
}

func (f *FetchAuthor) Run() {
  f.Person = &models.Person{}

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    (*f).Err = err
    return
  }
  defer db.Close()

  _, host, err := helpers.ParseAuthor(f.Author)
  if err != nil {
    revel.ERROR.Println(err)
    (*f).Err = err
    return
  }

  // skip updating person already known
  err = db.Where("author = ?", f.Author).First(f.Person).Error
  if err == nil { return }

  // add host to pod list
  pod := models.Pod{Host: host}
  if err := db.FirstOrCreate(&pod).Error; err != nil {
    revel.ERROR.Println(err)
    (*f).Err = err
    return
  }

  webFinger := federation.WebFinger{
    Host: host, Handle: f.Author,
  }; err = webFinger.Discovery()
  if err != nil {
    revel.ERROR.Println(err)
    (*f).Err = err
    return
  }

  var hcard federation.Hcard
  for _, link := range webFinger.Json.Links {
    if link.Rel == federation.WebFingerHcard {
      if err = hcard.Fetch(link.Href); err != nil {
        revel.ERROR.Println(err)
        (*f).Err = err
        return
      }
    }
  }

  if hcard.Guid == "" || hcard.PublicKey == "" {
    (*f).Err = errors.New("Something went wrong! Hcard struct is empty.")
    return
  }
  revel.TRACE.Println("Fetched hcard", hcard)

  *f.Person = models.Person{
    Guid: hcard.Guid,
    Author: f.Author,
    SerializedPublicKey: hcard.PublicKey,
    PodID: pod.ID,
    Profile: models.Profile{
      Author: f.Author,
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

  err = db.Save(f.Person).Error
  if err != nil {
    revel.ERROR.Println(err)
    (*f).Err = err
    return
  }
  revel.TRACE.Println(f.Person, f.Person.Profile)
}
