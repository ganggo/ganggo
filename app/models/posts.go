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
  federation "gopkg.in/ganggo/federation.v0"
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
  Comments Comments `gorm:"ForeignKey:ShareableID";json:",omitempty"`
}

type Posts []Post

func (p Posts) Len() int { return len(p) }
func (p Posts) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Posts) Less(i, j int) bool {
  return p[i].UpdatedAt.After(p[j].UpdatedAt)
}

func (post *Post) AfterFind(db *gorm.DB) error {
  if structLoaded(post.Person.CreatedAt) {
    return nil
  }

  err := db.Model(post).Related(&post.Person).Error
  if err != nil {
    return err
  }

  return db.Preload("Comments").First(post).Error
}

func (p *Post) AfterCreate(db *gorm.DB) error {
  // batch insert doesn't work for gorm, yet
  // see https://github.com/jinzhu/gorm/issues/255
  tags, err := generateTags(p)
  if err == nil && len(tags) > 0 {
    for _, tag := range tags {
      var cnt int
      db.Where("name = ?", tag.Name).Find(&tag).Count(&cnt)
      // if tag already exists skip it
      // and create taggings only
      if cnt == 0 {
        err = db.Create(&tag).Error
        if err != nil {
          return err
        }
      } else {
        for _, shareable := range tag.ShareableTaggings {
          shareable.TagID = tag.ID
          err = db.Create(&shareable).Error
          if err != nil {
            return err
          }
        }
      }
    }
  } else if err != nil {
    return err
  }

  notify, err := generateNotifications(p)
  if err == nil && len(notify) > 0 {
    for _, n := range notify {
      err = db.Create(&n).Error
      if err != nil {
        return err
      }
    }
  }
  return err
}

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

func (posts *Posts) FindAllPublic(offset int) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Offset(offset).Limit(10).
    Where(`public = ?`, true).Order(`updated_at desc`)

  return query.Find(posts).Error
}

func (posts *Posts) FindAll(userID uint, offset int) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Offset(offset).Limit(10).
    Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where(`posts.public = ?`, true).
    Or(`posts.id = shareables.shareable_id
      and shareables.shareable_type = ?
      and shareables.user_id = ?`,
        ShareablePost, userID,
    ).Order("posts.updated_at desc").Find(posts).Error
}

func (posts *Posts) FindAllPrivate(userID uint, offset int) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Offset(offset).Limit(10).
    Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where(`posts.id = shareables.shareable_id
      and shareables.shareable_type = ?
      and shareables.user_id = ?`, ShareablePost, userID,
    ).Order(`posts.updated_at desc`).Find(posts).Error
}

func (posts *Posts) FindAllByUserAndPersonID(user User, personID uint, offset int) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Offset(offset).Limit(10).
    Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where(`posts.public = ? and person_id = ?`, true, personID)

  if user.SerializedPrivateKey != "" {
    query = query.Or(`posts.id = shareables.shareable_id
      and shareables.shareable_type = ?
      and shareables.user_id = ?
      and person_id = ?`, ShareablePost, user.ID, personID)
  }
  return query.Order(`posts.updated_at desc`).Find(posts).Error
}


func (post *Post) FindByID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(post, id).Error
}

func (post *Post) FindByGuid(guid string) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("guid = ?", guid).Find(post).Error
}

func (post *Post) FindByGuidUser(guid string, user User) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where(`posts.public = true and guid = ?`, guid)

  if user.SerializedPrivateKey != "" {
    query = query.Or(`posts.id = shareables.shareable_id
        and shareables.shareable_type = ?
        and shareables.user_id = ?
        and guid = ?`, ShareablePost, user.ID, guid)
  }

  return query.Find(post).Error
}
