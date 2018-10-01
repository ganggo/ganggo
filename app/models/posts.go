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
  "git.feneas.org/ganggo/gorm"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/federation"
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
  ProviderName string `gorm:"size:191"`
  RootGuid *string `gorm:"size:187"`
  RootPersonID uint
  Protocol federation.Protocol `gorm:"size:4"`

  Person Person `gorm:"ForeignKey:PersonID" json:",omitempty"`
  Comments Comments `gorm:"ForeignKey:ShareableID" json:",omitempty"`
  Photos Photos
}

type Posts []Post

// Model Interface Type
//   FetchID() uint
//   FetchGuid() string
//   FetchType() string
//   FetchPersonID() uint
//   FetchText() string
//   HasPublic() bool
//   IsPublic() bool
func (p Post) FetchID() uint { return p.ID }
func (p Post) FetchGuid() string { return p.Guid }
func (Post) FetchType() string { return ShareablePost }
func (p Post) FetchPersonID() uint { return p.PersonID }
func (p Post) FetchText() string { return p.Text }
func (Post) HasPublic() bool { return true }
func (p Post) IsPublic() bool { return p.Public }
// Model Interface Type

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

  err = db.Model(post).Related(&post.Photos).Error
  if err != nil {
    return err
  }

  return db.Preload("Comments").First(post).Error
}

func (p *Post) AfterSave(db *gorm.DB) error {
  err := p.AfterFind(db)
  if err != nil {
    return err
  }

  err = searchAndCreateTags(*p, db)
  if err != nil {
    return err
  }
  return checkForMentionsInText(*p)
}

func (p *Post) AfterDelete(db *gorm.DB) (err error) {
  // likes
  err = db.Where("shareable_id = ? and shareable_type = ?",
    p.ID, ShareablePost).Delete(Like{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // comments
  err = db.Where("shareable_id = ? and shareable_type = ?",
    p.ID, ShareablePost).Delete(Comment{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // visibilities
  err = db.Where("shareable_id = ? and shareable_type = ?",
    p.ID, ShareablePost).Delete(Visibility{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // aspect_visibilities
  err = db.Where("shareable_id = ? and shareable_type = ?",
    p.ID, ShareablePost).Delete(AspectVisibility{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // shareables
  err = db.Where("shareable_id = ? and shareable_type = ?",
    p.ID, ShareablePost).Delete(Shareable{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // shareable_taggings
  err = db.Where("shareable_id = ? and shareable_type = ?",
    p.ID, ShareablePost).Delete(ShareableTagging{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // photos
  err = db.Where("post_id = ?", p.ID).Delete(Photo{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  // notifications
  err = db.Where("shareable_guid = ? and shareable_type = ?",
    p.Guid, ShareablePost).Delete(Notification{}).Error
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  return
}

func (p *Post) Count() (count int) {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  db.Table("posts").Joins(
    `left join people on posts.person_id = people.id`,
  ).Where("people.user_id > 0").Count(&count)
  return
}

func (p *Post) Create() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  return db.Create(p).Error
}

func (p *Post) Delete() (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  if p.ID == 0 {
    // see http://jinzhu.me/gorm/crud.html#delete
    panic("Setting ID to zero will delete all post records")
  }
  return db.Delete(p).Error
}

func (p *Post) IsLocal() (user User, ok bool) {
  db, err := OpenDatabase()
  if err != nil {
    revel.WARN.Println(err)
    return user, false
  }
  defer db.Close()

  if p.Person.UserID > 0 {
    err = db.First(&user, p.Person.UserID).Error
    if err != nil {
      return user, false
    }
    return user, true
  }
  return user, false
}

func (posts *Posts) FindAllPublic(offset uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Offset(offset).Limit(10).
    Where(`public = ?`, true).Order(`created_at desc`)

  return query.Find(posts).Error
}

func (posts *Posts) FindAllPublicByPerson(person Person, offset uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Offset(offset).Limit(10).
    Where(`person_id = ? and public = ?`, person.ID, true).Order(`created_at desc`)

  return query.Find(posts).Error
}

func (posts *Posts) FindAll(userID, offset uint) (err error) {
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
    ).Order("posts.created_at desc").Find(posts).Error
}

func (posts *Posts) FindAllPrivate(userID, offset uint) (err error) {
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
    ).Order(`posts.created_at desc`).Find(posts).Error
}

func (posts *Posts) FindAllByUserAndPersonID(user User, personID, offset uint) (err error) {
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
  return query.Order(`posts.created_at desc`).Find(posts).Error
}

func (posts *Posts) FindAllByUserAndText(user User, text string, offset uint) (err error) {
  if text == "" {
    revel.AppLog.Debug("Skipping empty string search")
    return
  }

  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Offset(offset).Limit(10).
    Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where(`posts.public = ? and ?`, true, advancedColumnSearch("text", text))
  if user.SerializedPrivateKey != "" {
    query = query.Or(`posts.id = shareables.shareable_id
      and shareables.shareable_type = ?
      and shareables.user_id = ?
      and ?`, ShareablePost, user.ID, advancedColumnSearch("text", text),
    )
  }
  return query.Order(`posts.created_at desc`).Find(posts).Error
}

func (post *Post) Exists(id uint) bool {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error("Post.Exists", "err", err)
    return false
  }
  defer db.Close()

  return !db.Find(post, id).RecordNotFound()
}

func (post *Post) FindByID(id uint) (err error) { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(post, id).Error
}

func (post *Post) FindByIDAndUser(id uint, user User) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  query := db.Joins(`left join shareables on shareables.shareable_id = posts.id`).
    Where(`posts.public = true and posts.id = ?`, id)

  if user.SerializedPrivateKey != "" {
    query = query.Or(`posts.id = shareables.shareable_id
        and shareables.shareable_type = ?
        and shareables.user_id = ?
        and posts.id = ?`, ShareablePost, user.ID, id)
  }

  return query.Find(post).Error
}

func (post *Post) FindByGuid(guid string) (err error) { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("guid = ?", guid).Find(post).Error
}

func (post *Post) FindByGuidAndUser(guid string, user User) (err error) {
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
