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

import "time"

type Aspect struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Name string `gorm:"size:191"`
  UserID uint
  Default bool

  Memberships []AspectMembership `json:",omitempty"`
}

type Aspects []Aspect

type AspectMembership struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  AspectID uint
  PersonID uint
}

type AspectVisibility struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  ShareableID uint
  AspectID uint
  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  ShareableType string `gorm:"size:191"`
}

type AspectVisibilities []AspectVisibility

func (aspect *Aspect) Create() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(aspect).Error
}

func (visibility *AspectVisibility) Create() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(visibility).Error
}

func (visibility *AspectVisibility) FindByGuid(guid string) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  var post Post
  err = post.FindByGuid(guid)
  if err != nil {
    return err
  }

  return db.Where("shareable_id = ? and shareable_type = ?", post.ID, ShareablePost).Find(visibility).Error
}

func (aspect *Aspect) FindByID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.Find(aspect, id).Error
  if err != nil {
    return err
  }

  db.Model(aspect).Related(&aspect.Memberships)

  return
}

func (aspects *Aspects) FindByUserPersonID(userID, personID uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Table("aspects").
    Joins("left join aspect_memberships on aspect_memberships.aspect_id = aspects.ID").
    Where("aspects.user_id = ? and aspect_memberships.person_id = ?", userID, personID).
    Find(&aspects).Error
}

func (membership *AspectMembership) Create() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(membership).Error
}

func (membership *AspectMembership) Delete() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("aspect_id = ? and person_id = ?",
    membership.AspectID, membership.PersonID,
  ).Delete(membership).Error
}
