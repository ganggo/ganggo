package models
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
  "time"
  "github.com/jinzhu/gorm"
)

type EventStatus int

const (
  EventAccepted EventStatus = iota
  EventDeclined
  EventTentative
)

type Calendar struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  Name string `gorm:"size:191"`
  UserID uint
  Default bool

  Events CalendarEvents
}

type Calendars []Calendar

type CalendarEvent struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  Public bool
  PersonID uint
  Location string
  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Guid string `gorm:"size:191"`
  Summary string `gorm:"type:text"`
  Description string `gorm:"type:text"`
  Start time.Time
  End time.Time
  AllDay bool
  CalendarID uint

  Person Person
  Participations CalendarEventParticipations
}

type CalendarEvents []CalendarEvent

type CalendarEventParticipation struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  PersonID uint
  CalendarEventID uint
  Status EventStatus

  Person Person
}

type CalendarEventParticipations []CalendarEventParticipation

func (c *Calendar) AfterFind(db *gorm.DB) error {
  return db.Model(c).Related(&c.Events).Error
}

func (e *CalendarEvent) AfterFind(db *gorm.DB) error {
  err := db.Model(e).Related(&e.Participations).Error
  // if its a public event CalendarID can be zero
  // so ignore RecordNotFound errors
  if err != nil && err != gorm.ErrRecordNotFound {
    return err
  }
  return db.Model(e).Related(&e.Person).Error
}

func (c *Calendar) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(c).Error
}

func (e *CalendarEvent) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(e).Error
}

func (c *Calendars) FindByUser(user User) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ?", user.ID).Find(c).Error
}

func (c *Calendar) FindByUserAndID(user User, id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ? and id = ?", user.ID, id).Find(c).Error
}

func (c *Calendar) FindByUserAndName(user User, name string) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ? and name = ?", user.ID, name).Find(c).Error
}

func (e *CalendarEvent) FindByUserAndID(user User, id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.First(e, id).Error
  if err != nil {
    return err
  }
  // if it is public no need
  // for further verifiction
  if e.Public {
    return nil
  }
  return db.Joins(`left join shareables on shareables.shareable_id = calendar_events.id`).
    Where(`calendar_events.id = ?
      and shareables.shareable_type = ?
      and shareables.user_id = ?`, id, ShareableEvent, user.ID,
    ).Find(e).Error
}

func (e *CalendarEvents) FindAllPublic() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("public = ?", true).Find(e).Error
}
