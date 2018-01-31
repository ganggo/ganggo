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

type Calendar struct {
  *revel.Controller
}

func (c Calendar) Index(name string, page int) revel.Result {
  user, err := models.CurrentUser(c.Controller)
  if err != nil {
    c.Log.Error("Cannot find user", "error", err)
    return c.RenderError(err)
  }
  // on error the template will display
  // an option to create a new calendar
  var calendars models.Calendars
  calendars.FindByUser(user)

  c.ViewArgs["calendars"] = calendars
  c.ViewArgs["currentUser"] = user

  return c.RenderTemplate("calendar/index.html")
}

func (c Calendar) Public(page int) revel.Result {
  user, err := models.CurrentUser(c.Controller)
  if err != nil {
    c.Log.Error("Cannot find user", "error", err)
    return c.RenderError(err)
  }

  var events models.CalendarEvents
  err = events.FindAllPublic()
  if err != nil {
    c.Log.Error(err.Error())
    return c.RenderError(err)
  }

  c.ViewArgs["page"] = page
  c.ViewArgs["currentUser"] = user
  c.ViewArgs["calendar"] = struct{
    ID uint
    Events models.CalendarEvents
  }{0, events}

  return c.RenderTemplate("calendar/calendar.html")
}

func (c Calendar) Show(name string, page int) revel.Result {
  user, err := models.CurrentUser(c.Controller)
  if err != nil {
    c.Log.Error("Cannot find user", "error", err)
    return c.RenderError(err)
  }

  var calendar models.Calendar
  err = calendar.FindByUserAndName(user, name)
  if err != nil {
    c.Log.Error(err.Error())
    return c.RenderError(err)
  }

  c.ViewArgs["page"] = page
  c.ViewArgs["currentUser"] = user
  c.ViewArgs["calendar"] = calendar

  return c.RenderTemplate("calendar/calendar.html")
}

func (c Calendar) ShowEvent(id int) revel.Result {
  return c.NotFound("NA")
}
