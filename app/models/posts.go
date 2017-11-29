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
  federation "gopkg.in/ganggo/federation.v0"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Post struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  PersonID uint `gorm:"size:4"`
  Public bool
  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Guid string `gorm:"size:191"`
  Type string `gorm:"size:40"`
  Text string `gorm:"type:text"`
  ProviderName string
  RootGuid string
  RootHandle string
  LikesCount int `gorm:"size:4"`
  CommentsCount int `gorm:"size:4"`
  ResharesCount int `gorm:"size:4"`
  InteractedAt string

  Person Person `gorm:"ForeignKey:PersonID";json:",omitempty"`
  Comments []Comment `gorm:"ForeignKey:ShareableID";json:",omitempty"`
}

type Posts []Post

func (p *Post) Count() (count int, err error) {
  db, err := OpenDatabase()
  if err != nil {
    return -1, err
  }
  defer db.Close()

  db.Table("posts").Joins(
    `left join people on posts.person_id = people.id`,
  ).Where("people.user_id > 0").Count(&count)
  return
}

func (p *Post) Create(entity *federation.EntityStatusMessage, reshare bool) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  err = p.Cast(entity, reshare)
  if err != nil {
    return
  }

  return db.Create(p).Error
}

func (p *Post) Cast(entity *federation.EntityStatusMessage, reshare bool) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  var person Person
  err = db.Where("author = ?", entity.Author).First(&person).Error
  if err != nil {
    return
  }

  messageType := StatusMessage
  if reshare {
    messageType = Reshare
  }
  (*p).PersonID = person.ID
  (*p).Public = entity.Public
  (*p).Guid = entity.Guid
  (*p).RootGuid = entity.RootGuid
  (*p).RootHandle = entity.RootHandle
  (*p).Type = messageType
  (*p).Text = entity.Text
  (*p).ProviderName = entity.ProviderName

  return nil
}

func (p *Post) IsLocal() (User, bool) {
  return parentIsLocal(p.ID)
}

func (posts *Posts) FindAll(userID uint, offset int) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.Offset(offset).Limit(10).Table("posts").
    Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where("posts.public = true").
    Or(`posts.id = shareables.shareable_id
      and shareables.shareable_type = ?
      and shareables.user_id = ?`,
        ShareablePost, userID,
    ).Order("posts.updated_at desc").Find(posts).Error
  if err != nil {
    return err
  }
  return posts.addRelations(db)
}

func (posts *Posts) FindAllByPersonID(id uint, offset int) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.Offset(offset).Limit(10).
    Where("person_id = ?", id).
    Order("posts.updated_at desc").Find(posts).Error
  if err != nil {
    return
  }
  return posts.addRelations(db)
}


func (post *Post) FindByID(id uint, withRelations bool) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.Find(post, id).Error
  if err != nil {
    return
  }
  // add relations only if it is required
  if withRelations {
    return post.addRelations(db)
  }
  return
}

func (post *Post) FindByGuid(guid string) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.Where("guid = ?", guid).Find(post).Error
  if err != nil {
    return
  }
  return post.addRelations(db)
}

func (post *Post) FindByGuidUser(guid string, user User) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where(`posts.public = true`).Where("guid = ?", guid)

  if user.SerializedPrivateKey != "" {
    query = query.Or(`posts.id = shareables.shareable_id
        and shareables.shareable_type = ?
        and shareables.user_id = ?`, ShareablePost, user.ID).
      Where("guid = ?", guid)
  }

  err = query.Find(post).Error
  if err != nil {
    return
  }
  return post.addRelations(db)
}

func (posts *Posts) FindByTagName(name string, user User, offset int) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Offset(offset).Limit(10).
    Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where(`posts.public = true`).
    Where("text like ?", "%#"+name+"%")

  if user.SerializedPrivateKey != "" {
    query = query.Or(`posts.id = shareables.shareable_id
        and shareables.shareable_type = ?
        and shareables.user_id = ?`, ShareablePost, user.ID).
      Where("text like ?", "%#"+name+"%")
  }

  err = query.Order("posts.updated_at desc").Find(posts).Error
  if err != nil {
    return err
  }
  return posts.addRelations(db)
}

func (posts *Posts) addRelations(db *gorm.DB) error {
  for index := range *posts {
    var p *Post = &(*posts)[index]
    err := p.addRelations(db)
    if err != nil {
      return err
    }
  }
  return nil
}

func (post *Post) addRelations(db *gorm.DB) error {
  err := db.Model(post).Related(&post.Person).Error
  if err != nil {
    return err
  }
  err = db.Model(&post.Person).Related(&post.Person.Profile).Error
  if err != nil {
    return err
  }
  err = db.Preload("Comments").First(post).Error
  if err != nil {
    return err
  }
  for index := range post.Comments {
    var c *Comment = &post.Comments[index]
    err = c.addRelations(db)
    if err != nil {
      return err
    }
  }
  return nil
}
