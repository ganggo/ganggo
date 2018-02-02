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
  "strings"
  "github.com/jinzhu/gorm"
  "sort"
  "errors"
  "time"
)

type User struct {
  gorm.Model

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Username string `gorm:"size:191"`
  Email string `gorm:"size:191"`
  SerializedPrivateKey string `gorm:"type:text" json:"-" xml:"-"`
  EncryptedPassword string `json:"-" xml:"-"`

  PersonID uint
  Person Person `gorm:"ForeignKey:PersonID"`

  Aspects []Aspect `gorm:"AssociationForeignKey:UserID"`
}

type UserStream struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  UserID uint
  Name string `gorm:"size:191"`

  Tags string
  People string
  Expression string

  User User `json:"-"`
}

type UserStreams []UserStream

func (user *User) AfterCreate(tx *gorm.DB) error {
  return tx.Model(&user.Person).Update("user_id", user.ID).Error
}

func (user *User) AfterFind(db *gorm.DB) error {
  if structLoaded(user.Person.CreatedAt) {
    return nil
  }

  err := db.Model(user).Related(&user.Person).Error
  if err != nil {
    return err
  }

  return db.Model(user).Related(&user.Aspects).Error
}

func (user *User) FindByID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(user, id).Error
}

func (user *User) FindByUsername(name string) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("username = ?", name).Find(user).Error
}

func (user *User) Count() (count int, err error) {
  db, err := OpenDatabase()
  if err != nil {
    return -1, err
  }
  defer db.Close()

  db.Table("users").Count(&count)
  return
}

func (user *User) Notify(model Model) error {
  // do not send notification for your own activity
  if user.Person.ID == model.FetchPersonID() {
    return nil
  }

  notify := Notification{
    ShareableType: model.FetchType(),
    ShareableGuid: model.FetchGuid(),
    UserID: user.ID,
    PersonID: model.FetchPersonID(),
    Unread: true,
  }
  return notify.Create()
}

func (stream *UserStream) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  return db.Create(stream).Error
}

func (stream *UserStream) FindByName(name string) error { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  return db.Where("name = ?", name).Find(stream).Error
}

func (stream *UserStream) FetchPosts(posts *Posts, offset int) error { BACKEND_ONLY()
  tagNames := strings.Split(stream.Tags, ",")
  people := strings.Split(stream.People, ",")

  for _, name := range tagNames {
    var tag Tag
    err := tag.FindByName(name, stream.User, offset)
    if err != nil {
      if err == gorm.ErrRecordNotFound {
        continue
      }
      return err
    }
    for _, tagging := range tag.ShareableTaggings {
      *posts = append(*posts, tagging.Post)
    }
  }

  for _, author := range people {
    var person Person
    err := person.FindByAuthor(author)
    if err != nil {
      if err == gorm.ErrRecordNotFound {
        continue
      }
      return err
    }
    var authorPosts Posts
    err = authorPosts.FindAllByUserAndPersonID(
      stream.User, person.ID, offset)
    if err != nil && err != gorm.ErrRecordNotFound {
      return err
    }
    *posts = append(*posts, authorPosts...)
  }

  var expressionPosts Posts
  err := expressionPosts.FindAllByUserAndText(
    stream.User, stream.Expression, offset)
  if err != nil && err != gorm.ErrRecordNotFound {
    return err
  }
  *posts = append(*posts, expressionPosts...)

  sort.Sort(*posts)

  if len(*posts) < offset {
    *posts = (*posts)[:offset-1]
  }

  return nil
}

func (streams *UserStreams) FindByUser(user User) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ?", user.ID).Find(streams).Error
}

func (stream *UserStream) Delete() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  if stream.ID == 0 || stream.UserID == 0 {
    // NOTE ID being zero will delete ALL entries
    return errors.New("Cannot delete user stream without ID and UserID")
  }
  return db.Delete(stream).Error
}
