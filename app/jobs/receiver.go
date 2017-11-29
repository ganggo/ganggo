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
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  federation "gopkg.in/ganggo/federation.v0"
  "strings"
  //"net/url"
  //"fmt"
  //"encoding/xml"
)

type Receiver struct {
  Entity federation.Entity
  Guid string
}

func (r *Receiver) Run() {
  db, err := models.OpenDatabase()
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  // TODO signature verification for entities missing !!!
  switch name := r.Entity.Type; name {
  case federation.Retraction:
    r.Retraction(r.Entity.Data.(federation.EntityRetraction))
  case federation.Profile:
    var (
      profile models.Profile
      profileEntity = r.Entity.Data.(federation.EntityProfile)
    )

    revel.TRACE.Println("Found a profile entity", profileEntity)

    insert := db.Where("author = ?",
      profileEntity.Author,
    ).First(&profile).RecordNotFound()

    err := profile.Cast(&profileEntity)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }

    if !strings.HasPrefix(profile.ImageUrl, "http") {
      _, host, err := helpers.ParseAuthor(profile.Author)
      if err != nil {
        revel.ERROR.Println(err)
        return
      }
      url := "https://" + host
      profile.ImageUrl = url + profile.ImageUrl
      profile.ImageUrlMedium = url + profile.ImageUrlMedium
      profile.ImageUrlSmall = url + profile.ImageUrlSmall
    }

    if insert {
      err = db.Create(&profile).Error
      if err != nil {
        revel.ERROR.Println(err, profile)
        return
      }
    } else {
      err = db.Save(&profile).Error
      if err != nil {
        revel.ERROR.Println(err, profile)
        return
      }
    }
  case federation.StatusMessage, federation.Reshare:
    var (
      post models.Post
      user models.Person
      person models.Person
      message = r.Entity.Data.(federation.EntityStatusMessage)
    )

    revel.TRACE.Println("Found a status_message entity")

    fetch := FetchAuthor{Author: message.Author}
    fetch.Run()

    var reshare bool
    if name == federation.Reshare {
      reshare = true
    }

    err := post.Cast(&message, reshare)
    if err != nil {
      revel.WARN.Println(err)
      return
    }

    err = db.Create(&post).Error
    if err != nil {
      revel.TRACE.Println(err)
      // XXX ignore if it fails cause
      // if multiple user receive the
      // same private status message you
      // need a shareable item for everyone
      // even if the post already exists
      //return
    }

    err = db.Where("guid = ?", r.Guid).First(&person).Error
    if err != nil {
      if r.Guid != "" {
        revel.ERROR.Println(err, r.Guid)
      }
      // otherwise just return it is a public request
      return
    }

    err = db.Find(&user, person.UserID).Error
    if err != nil {
      revel.ERROR.Println(err)
      return
    }

    shareable := models.Shareable{
      ShareableID: post.ID,
      ShareableType: models.ShareablePost,
      UserID: user.ID,
    }
    err = db.Create(&shareable).Error
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
  case federation.Comment:
    r.Comment(r.Entity.Data.(federation.EntityComment))
  case federation.Like:
    r.Like(r.Entity.Data.(federation.EntityLike))
  case federation.Contact:
    r.Contact(r.Entity.Data.(federation.EntityContact))
  default:
    revel.ERROR.Println("No matching entity found for", name)
    return
  }
}
