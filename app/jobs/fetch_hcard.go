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
  "strings"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  "github.com/PuerkitoBio/goquery"
  federation "gopkg.in/ganggo/federation.v0"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type FetchHcard struct {
  Person *models.Person
  Err error
}

func (f *FetchHcard) Run() {
  var profile models.Profile

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    (*f).Err = err
    return
  }
  defer db.Close()

  _, host, err := helpers.ParseDiasporaHandle(f.Person.DiasporaHandle)
  if err != nil {
    revel.ERROR.Println(err)
    (*f).Err = err
    return
  }

  profile.DiasporaHandle = f.Person.DiasporaHandle
  resp, err := federation.FetchHtml("GET",
    host + "/hcard/users/" + f.Person.Guid, nil,
  )
  if err != nil {
    (*f).Err = err
    return
  }
  doc, err := goquery.NewDocumentFromResponse(resp)
  if err != nil {
    (*f).Err = err
    return
  }

  notFound := db.Where("diaspora_handle = ?",
    profile.DiasporaHandle).First(&profile).RecordNotFound()
  profile.PersonID = f.Person.ID

  doc.Find(".entity_first_name").Each(
  func(i int, s *goquery.Selection) {
    profile.FirstName = s.Find("span").Text()
  })
  doc.Find(".entity_family_name").Each(
  func(i int, s *goquery.Selection) {
    profile.LastName = s.Find("span").Text()
  })
  doc.Find(".entity_searchable").Each(
  func(i int, s *goquery.Selection) {
    var searchable bool = false
    if s.Find("span").Text() == "true" {
      searchable = true
    }
    profile.Searchable = searchable
  })

  // XXX do we want to support http as well ?
  host = "https://" + host

  doc.Find(".entity_photo").Each(
  func(i int, s *goquery.Selection) {
    nodes := s.Find("img")
    for _, node := range nodes.Nodes {
      for _, attr := range node.Attr {
        if attr.Key == "src" {
          value := attr.Val
          if !strings.HasPrefix(value, "http") {
            value = host + value
          }
          profile.ImageUrl = value
        }
      }
    }
  })
  doc.Find(".entity_photo_medium").Each(
  func(i int, s *goquery.Selection) {
    nodes := s.Find("img")
    for _, node := range nodes.Nodes {
      for _, attr := range node.Attr {
        if attr.Key == "src" {
          value := attr.Val
          if !strings.HasPrefix(value, "http") {
            value = host + value
          }
          profile.ImageUrlMedium = value
        }
      }
    }
  })
  doc.Find(".entity_photo_small").Each(
  func(i int, s *goquery.Selection) {
    nodes := s.Find("img")
    for _, node := range nodes.Nodes {
      for _, attr := range node.Attr {
        if attr.Key == "src" {
          value := attr.Val
          if !strings.HasPrefix(value, "http") {
            value = host + value
          }
          profile.ImageUrlSmall = value
        }
      }
    }
  })

  if notFound {
    err = db.Create(&profile).Error
    if err != nil {
      (*f).Err = err
    }
  } else {
    err = db.Save(&profile).Error
    if err != nil {
      (*f).Err = err
    }
  }
}
