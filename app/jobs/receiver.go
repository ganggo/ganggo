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
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
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
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  // TODO signature verification for entities missing !!!
  switch {
  case r.Entity.Post.Request != nil:
    var contact models.Contact

    revel.TRACE.Println("Found a request entity")

    err := contact.Cast(r.Entity.Post.Request)
    if err != nil {
      revel.WARN.Println(err)
      return
    }

    err = db.Create(&contact).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
  case r.Entity.Post.Retraction != nil:
    var person models.Person

    revel.TRACE.Println("Found a retraction entity")

    retract := r.Entity.Post.Retraction
    err := db.Where("guid = ?", retract.PostGuid).First(&person).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }

    // TODO there are probably more types
    // do not delete on every retraction some contacts
    return

    for _, contact := range person.Contacts {
      err = db.Delete(&contact).Error
      if err != nil {
        revel.WARN.Println(err)
      }
    }
    return
  case r.Entity.Post.Profile != nil:
    var profile models.Profile

    revel.TRACE.Println("Found a profile entity")

    insert := db.Where("diaspora_handle = ?",
      r.Entity.Post.Profile.DiasporaHandle,
    ).First(&profile).RecordNotFound()

    err := profile.Cast(r.Entity.Post.Profile)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }

    if !strings.HasPrefix(profile.ImageUrl, "http") {
      _, host, err := helpers.ParseDiasporaHandle(profile.DiasporaHandle)
      if err != nil {
        revel.ERROR.Println(err)
        return
      }
      url := "http://" + host
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
  case r.Entity.Post.StatusMessage != nil || r.Entity.Post.Reshare != nil:
    var (
      post models.Post
      user models.Person
      person models.Person
    )

    revel.TRACE.Println("Found a status_message entity")

    err := post.Cast(
      r.Entity.Post.StatusMessage,
      r.Entity.Post.Reshare,
    )
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
  case r.Entity.Post.Comment != nil:
    r.Comment()
  case r.Entity.Post.Like != nil:
    var like models.Like

    revel.TRACE.Println("Found a like entity")

    err := like.Cast(r.Entity.Post.Like)
    if err != nil {
      revel.WARN.Println(err)
      return
    }

    err = db.Create(&like).Error
    if err != nil {
      revel.WARN.Println(err)
      return
    }
  case r.Entity.Post.RelayableRetraction != nil:
    revel.TRACE.Println("Found a relayable_retraction")
    retraction := r.Entity.Post.RelayableRetraction

    switch {
    case strings.EqualFold("post", retraction.TargetType):
      var post models.Post
      err := db.Where("guid = ?", retraction.TargetGuid).First(&post).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
      err = db.Delete(&post).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
    case strings.EqualFold("like", retraction.TargetType):
      var like models.Like
      err := db.Where("guid = ?", retraction.TargetGuid).First(&like).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
      err = db.Delete(&like).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
    case strings.EqualFold("comment", retraction.TargetType):
      var comment models.Comment
      err := db.Where("guid = ?", retraction.TargetGuid).First(&comment).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
      err = db.Delete(&comment).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
    default:
      revel.ERROR.Println("Unknown retraction:", retraction)
      return
    }
  case r.Entity.Post.SignedRetraction != nil:
    revel.TRACE.Println("Found a signed_retraction")
    retraction := r.Entity.Post.SignedRetraction

    switch {
    case strings.EqualFold("post", retraction.TargetType):
      var post models.Post
      err := db.Where("guid = ?", retraction.TargetGuid).First(&post).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
      err = db.Delete(&post).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
    case strings.EqualFold("like", retraction.TargetType):
      var like models.Like
      err := db.Where("guid = ?", retraction.TargetGuid).First(&like).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
      err = db.Delete(&like).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
    case strings.EqualFold("comment", retraction.TargetType):
      var comment models.Comment
      err := db.Where("guid = ?", retraction.TargetGuid).First(&comment).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
      err = db.Delete(&comment).Error
      if err != nil {
        revel.WARN.Println(err)
        return
      }
    default:
      revel.ERROR.Println("Unknown retraction:", retraction)
      return
    }
  default:
    revel.ERROR.Println("No matching entity found! Ignoring it..")
    return
  }
}
