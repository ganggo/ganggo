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
  "github.com/revel/revel"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Database struct {
  Driver string
  Url string
}

const (
  Reshare = "Reshare"
  StatusMessage = "StatusMessage"
  ShareablePost = "Post"
)

var DB Database

func InitDB() {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    panic("failed to connect database" + DB.Driver + DB.Url + err.Error())
  }
  defer db.Close()

  // use revel logger
  db.SetLogger(gorm.Logger{revel.TRACE})

  post := &Post{}
  db.Model(post).AddUniqueIndex("index_posts_on_guid", "guid")
  //db.Model(post).AddUniqueIndex("index_posts_on_person_id_and_root_guid", "person_id", "root_guid")
  db.Model(post).AddIndex("index_posts_on_person_id", "person_id")
  db.Model(post).AddIndex("index_posts_on_id_and_type_and_created_at", "id", "type", "created_at")
  //db.Model(post).AddIndex("index_posts_on_root_guid", "root_guid")
  db.Model(post).AddIndex("index_posts_on_id_and_type", "id", "type")
  db.AutoMigrate(post)

  comment := &Comment{}
  db.Model(comment).AddUniqueIndex("index_comments_on_guid", "guid")
  db.Model(comment).AddIndex("index_comments_on_person_id", "person_id")
  db.Model(comment).AddIndex("index_comments_on_shareable_id_and_shareable_type", "shareable_id", "shareable_type")
  db.AutoMigrate(comment)

  signature := &CommentSignature{}
  db.Model(signature).AddUniqueIndex("index_comment_signatures_on_comment_id", "comment_id")
  db.AutoMigrate(signature)

  contact := &Contact{}
  db.Model(contact).AddUniqueIndex("index_contacts_on_user_id_and_person_id", "user_id", "person_id")
  db.Model(contact).AddIndex("index_contacts_on_person_id", "person_id")
  db.AutoMigrate(contact)

  profile := &Profile{}
  db.Model(profile).AddUniqueIndex("index_profiles_on_person_id", "person_id")
  db.Model(profile).AddUniqueIndex("index_profiles_on_diaspora_handle", "diaspora_handle")
  db.Model(profile).AddIndex("index_profiles_on_full_name_and_searchable", "full_name", "searchable")
  db.Model(profile).AddIndex("index_profiles_on_full_name", "full_name")
  db.AutoMigrate(profile)

  person := &Person{}
  db.Model(person).AddUniqueIndex("index_people_on_diaspora_handle", "diaspora_handle")
  db.Model(person).AddUniqueIndex("index_people_on_guid", "guid")
  // NOTE not every person is a local user
  //db.Model(person).AddUniqueIndex("index_people_on_user_id", "user_id")
  db.Model(person).AddIndex("people_pod_id_fk", "pod_id")
  db.AutoMigrate(person)

  session := &Session{}
  db.Model(session).AddUniqueIndex("index_session_on_token_and_user_id", "token", "user_id")
  db.Model(session).AddIndex("index_session_on_token", "token")
  db.Model(session).AddIndex("index_session_on_user_id", "user_id")
  db.AutoMigrate(session)

  user := &User{}
  db.Model(user).AddUniqueIndex("index_users_on_username", "username")
  //db.Model(user).AddUniqueIndex("index_users_on_email", "email")
  //db.Model(user).AddUniqueIndex("index_users_on_authentication_token", "authentication_token")
  db.AutoMigrate(user)

  like := &Like{}
  db.Model(like).AddUniqueIndex("index_likes_on_target_id_and_person_id_and_target_type", "target_id", "person_id", "target_type")
  db.Model(like).AddUniqueIndex("index_likes_on_guid", "guid")
  db.Model(like).AddIndex("likes_person_id_fk", "person_id")
  db.Model(like).AddIndex("index_likes_on_post_id", "target_id")
  db.AutoMigrate(like)

  pod := &Pod{}
  db.Model(pod).AddUniqueIndex("index_pods_on_host", "host")
  db.AutoMigrate(pod)

  shareable := &Shareable{}
  db.Model(shareable).AddUniqueIndex("shareable_and_user_id", "shareable_id", "shareable_type", "user_id")
  db.Model(shareable).AddIndex("index_post_visibilities_on_post_id", "shareable_id")
  db.Model(shareable).AddIndex("index_share_visibilities_on_user_id", "user_id")
  db.Model(shareable).AddIndex("shareable_and_hidden_and_user_id", "shareable_id", "shareable_type", "hidden", "user_id")
  db.AutoMigrate(shareable)

  aspect := &Aspect{}
  //db.Model(aspect).AddIndex("index_aspects_on_user_id_and_contacts_visible", "user_id", "contacts_visible")
  db.Model(aspect).AddIndex("index_aspects_on_user_id", "user_id")
  db.AutoMigrate(aspect)

  aspectMembership := &AspectMembership{}
  db.Model(aspectMembership).AddUniqueIndex("index_aspect_memberships_on_aspect_id_and_person_id", "aspect_id", "person_id")
  db.Model(aspectMembership).AddIndex("index_aspect_memberships_on_aspect_id", "aspect_id")
  db.Model(aspectMembership).AddIndex("index_aspect_memberships_on_contact_id", "person_id")
  db.AutoMigrate(aspectMembership)

  aspectVisibility := &AspectVisibility{}
  db.Model(aspectVisibility).AddIndex("index_aspect_visibilities_on_aspect_id", "aspect_id")
  db.Model(aspectVisibility).AddIndex("shareable_and_aspect_id", "shareable_id", "shareable_type", "aspect_id")
  db.Model(aspectVisibility).AddIndex("index_aspect_visibilities_on_shareable_id_and_shareable_type", "shareable_id", "shareable_type")
  db.AutoMigrate(aspectVisibility)
}

func GetCurrentUser(token string) (user User, err error) {
  db, err := gorm.Open(DB.Driver, DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return user, err
  }
  defer db.Close()

  var session Session
  err = db.Where("token = ?", token).First(&session).Error
  if err != nil {
    revel.ERROR.Println(err)
    return user, err
  }

  err = db.First(&user, session.UserID).Error
  if err != nil {
    revel.ERROR.Println(err)
    return user, err
  }
  db.Model(&user).Related(&user.Person, "Person")
  db.Model(&user).Related(&user.Aspects)
  return
}
