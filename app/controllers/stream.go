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
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
)

type Stream struct {
  *revel.Controller
}

func (s Stream) Index(view string, page uint) revel.Result {
  var posts models.Posts
  var offset uint = ((page - 1) * 10)

  user, err := models.CurrentUser(s.Controller)
  if err == nil {
    s.ViewArgs["currentUser"] = user
  }

  if view == "public" {
    err := posts.FindAllPublic(offset)
    if err != nil {
      s.Log.Error("Cannot find posts", "error", err)
      return s.RenderError(err)
    }
  } else {
    user, err := models.CurrentUser(s.Controller)
    if err != nil {
      s.Log.Error("Cannot fetch current user", "error", err)
      return s.Redirect(User.Login)
    }
    s.ViewArgs["currentUser"] = user

    if view == "private" {
      err := posts.FindAllPrivate(user.ID, offset)
      if err != nil {
        s.Log.Error("Cannot find posts", "error", err)
        return s.RenderError(err)
      }
    } else {
      var userStream models.UserStream
      err := userStream.FindByName(view)
      if err != nil {
        s.Log.Error("Cannot find stream", "view", view, "error", err)
        return s.Redirect("/stream?view=private")
      }
      err = userStream.FetchPosts(&posts, offset)
      if err != nil {
        s.Log.Error("Cannot find posts", "error", err)
        return s.RenderError(err)
      }
    }
  }

  s.ViewArgs["title"] = revel.MessageFunc(
    s.Request.Locale, "stream.title",
  )
  if page <= 0 { page = 1 }
  s.ViewArgs["page"] = page
  s.ViewArgs["posts"] = posts
  s.ViewArgs["view"] = view

  return s.RenderTemplate("stream/index.html")
}

func (s Stream) IndexStreams() revel.Result {
  user, err := models.CurrentUser(s.Controller)
  if err != nil {
    s.Log.Error("Cannot find user", "error", err)
    return s.RenderError(err)
  }

  var streams models.UserStreams
  err = streams.FindByUser(user)
  if err != nil {
    s.Log.Error("Cannot find user streams", "error", err)
    return s.RenderError(err)
  }

  s.ViewArgs["userStreams"] = streams
  s.ViewArgs["currentUser"] = user

  return s.RenderTemplate("user/streams.html")
}
