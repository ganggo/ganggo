package controllers
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
  "net/http"
  "github.com/revel/revel"
  "github.com/ganggo/ganggo/app/models"
  "github.com/ganggo/ganggo/app/jobs"
  "github.com/ganggo/federation"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Stream struct {
  *revel.Controller
}

func (s Stream) Index() revel.Result {
  var offset int

  s.Params.Bind(&offset, "offset")
  posts := s.IndexPosts(offset)

  if offset > 0 {
    return s.RenderJson(posts)
  }

  s.RenderArgs["posts"] = posts

  return s.Render()
}

func (s *Stream) IndexPosts(offset int) (posts []models.Post) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    (*s).Response.Status = http.StatusInternalServerError
    revel.WARN.Println(err)
    return
  }
  defer db.Close()

  err = db.Offset(offset).Limit(10).Table("posts").
    Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where("posts.public = true").
    Or(`posts.ID = shareables.shareable_id and shareables.shareable_type = ?`,
      models.ShareablePost,
    ).
    Order("posts.updated_at desc").
    Find(&posts).
    Error
  if err != nil {
    (*s).Response.Status = http.StatusInternalServerError
    revel.WARN.Println(err)
    return
  }

  // load comments
  for i, post := range posts {
    err = db.Where("shareable_id = ?", post.ID).Find(&(posts[i].Comments)).Error
    if err != nil {
      (*s).Response.Status = http.StatusInternalServerError
      revel.WARN.Println(err)
      return
    }
  }
  return
}

func (s Stream) Create() revel.Result {
  var post, comment, parentGuid string
  var (
    user models.User
    session models.Session
  )

  s.Params.Bind(&post, "post")
  s.Params.Bind(&comment, "comment")
  s.Params.Bind(&parentGuid, "parent_guid")

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    s.Response.Status = http.StatusInternalServerError
    revel.WARN.Println(err)
    return s.Render()
  }
  defer db.Close()

  err = db.Where("token = ?", s.Session["TOKEN"]).First(&session).Error
  if err != nil {
    s.Response.Status = http.StatusInternalServerError
    revel.ERROR.Println(err)
    return s.Render()
  }

  err = db.First(&user, session.UserID).Error
  if err != nil {
    s.Response.Status = http.StatusInternalServerError
    revel.ERROR.Println(err)
    return s.Render()
  }

  dispatcher := jobs.Dispatcher{User: user}
  if post != "" {
    dispatcher.Message = federation.EntityStatusMessage{
      RawMessage: post,
    }
  } else if comment != "" {
    dispatcher.Message = federation.EntityComment{
      ParentGuid: parentGuid,
      Text: comment,
    }
  }
  go dispatcher.Run()

  return s.Redirect(Stream.Index)
}
