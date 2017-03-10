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
  "encoding/base64"
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
  var (
    profile models.Profile
    person models.Person
  )

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    (*f).Err = err
    return
  }
  defer db.Close()

  err = db.Where("diaspora_handle = ?", (*f).Author).First(&person).Error
  if err != nil {
    revel.TRACE.Println("No author with name", (*f).Author, "known to the db")

    // set diaspora handle
    person.DiasporaHandle = (*f).Author
    _, host, err := helpers.ParseDiasporaHandle((*f).Author)
    if err != nil {
      revel.ERROR.Println(err)
      (*f).Err = err
      return
    }

    webFinger := federation.WebFinger{
      Host: host,
      Handle: (*f).Author,
    }
    err = webFinger.Discovery()
    if err != nil {
      revel.ERROR.Println(err)
      (*f).Err = err
      return
    }

    for _, link := range webFinger.Xrd.Links {
      if link.Rel == "diaspora-public-key" {
        key, err := base64.StdEncoding.DecodeString(link.Href)
        if err != nil {
          revel.ERROR.Println(err)
          (*f).Err = err
          return
        }
        // set public key
        person.SerializedPublicKey = string(key)
      }
      if link.Rel == "http://joindiaspora.com/guid" {
        // set guid
        person.Guid = link.Href
      }
    }
    revel.TRACE.Println(person)

    err = db.Create(&person).Error
    if err != nil {
      revel.ERROR.Println(err)
      (*f).Err = err
      return
    }
  }

  err = db.Find(&profile, person.ID).Error
  if err != nil {
    revel.TRACE.Println("No profile for", (*f).Author, "available! Fetching..")

    fetchHcard := FetchHcard{
      Person: &person,
    }
    fetchHcard.Run()
    (*f).Err = fetchHcard.Err
  }
  (*f).Person = &person
}
