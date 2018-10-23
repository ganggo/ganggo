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
  "git.feneas.org/ganggo/gorm"
  "git.feneas.org/ganggo/federation"
  "git.feneas.org/ganggo/ganggo/app/helpers"
)

type SchemaMigration struct {
  Commit string `gorm:"size:191"`
}

type SchemaMigrations []SchemaMigration

func migrateSchema(db *gorm.DB) error {
  var commit string
  var structMigrations SchemaMigrations
  var migrations map[string]string = make(map[string]string)

  err := db.Find(&structMigrations).Error
  if err != nil {
    return err
  }

  for _, m := range structMigrations {
    migrations[m.Commit] = m.Commit
  }
  structMigrations = SchemaMigrations{}

  //// Migrations Start ////

  // related to https://git.feneas.org/ganggo/ganggo/merge_requests/77
  commit = "https://git.feneas.org/ganggo/ganggo/merge_requests/77"
  if _, ok := migrations[commit]; !ok {
    advancedColumnRename(db.Model(UserSetting{}), "key", "setting_key", "int(11)")
    structMigrations = append(structMigrations, SchemaMigration{Commit: commit})
  }

  // related to https://git.feneas.org/ganggo/ganggo/merge_requests/76
  commit = "https://git.feneas.org/ganggo/ganggo/merge_requests/76"
  if _, ok := migrations[commit]; !ok {
    var users Users
    err := db.Find(&users).Error
    if err != nil {
      return err
    }
    for _, user := range users {
      token, err := helpers.Token()
      if err != nil {
        return err
      }
      setting := UserSetting{
        UserID: user.ID,
        Key: UserSettingTelegramVerified,
        Value: token,
      }
      err = setting.Update()
      if err != nil {
        return err
      }
    }
    structMigrations = append(structMigrations, SchemaMigration{Commit: commit})
  }

  // related to https://git.feneas.org/ganggo/ganggo/merge_requests/55
  commit = "https://git.feneas.org/ganggo/ganggo/merge_requests/55"
  if _, ok := migrations[commit]; !ok {
    // set all existing posts to diaspora protocol
    protocol := map[string]interface{}{"protocol": federation.DiasporaProtocol}
    db.Model(Post{}).Updates(protocol)
    db.Model(Comment{}).Updates(protocol)
    db.Model(Like{}).Updates(protocol)
    db.Model(Pod{}).Updates(protocol)
    // delete unnecessary columns
    db.Model(Person{}).DropColumn("fetch_status")
    db.Model(Contact{}).DropColumn("receiving")
    db.Model(Pod{}).DropColumn("helo")
    db.Model(Post{}).DropColumn("likes_count")
    db.Model(Post{}).DropColumn("comments_count")
    db.Model(Post{}).DropColumn("reshares_count")
    db.Model(Post{}).DropColumn("interacted_at")
    db.Model(Profile{}).DropColumn("image_url_small")
    db.Model(Profile{}).DropColumn("image_url_medium")
    structMigrations = append(structMigrations, SchemaMigration{Commit: commit})
  }

  // related to ganggo/ganggo@d14d3e2e32730355a67b6cd418c94176914e27c0
  // see app/models/sessions.go for details
  commit = "d14d3e2e32730355a67b6cd418c94176914e27c0"
  if _, ok := migrations[commit]; !ok {
    db.Model(Session{}).DropColumn("id")
    db.Model(Session{}).DropColumn("updated_at")
    structMigrations = append(structMigrations, SchemaMigration{Commit: commit})
  }

  // related to ganggo/ganggo@0f94958c6b7f727c43031d756ae01d62d1467b74
  commit = "0f94958c6b7f727c43031d756ae01d62d1467b74"
  if _, ok := migrations[commit]; !ok {
    db.Model(Post{}).Where("root_guid = ?", "").Update("root_guid", gorm.Expr("NULL"))
    db.Model(Post{}).DropColumn("root_handle")
    advancedColumnModify(db.Model(Post{}), "root_guid", "varchar(187)")
    structMigrations = append(structMigrations, SchemaMigration{Commit: commit})
  }

  // related to ganggo/ganggo@2aec1bdfd61cfca7723b94cef3b09719cfb8e6f3
  commit = "2aec1bdfd61cfca7723b94cef3b09719cfb8e6f3"
  if _, ok := migrations[commit]; !ok {
    advancedColumnModify(db.Model(Aspect{}), "name", "varchar(191)")
    advancedColumnModify(db.Model(Pod{}), "host", "varchar(191)")
    structMigrations = append(structMigrations, SchemaMigration{Commit: commit})
  }

  //// Migrations End ////

  for _, m := range structMigrations {
    err = db.Create(&m).Error
    if err != nil {
      return err
    }
    revel.AppLog.Debug("Migration applied", "commit", m.Commit)
  }
  return nil
}

func loadSchema(db *gorm.DB) {
  schemaMigration := &SchemaMigration{}
  db.Model(schemaMigration).AddUniqueIndex("index_schema_migrations_on_commit", "commit")
  db.AutoMigrate(schemaMigration)

  post := &Post{}
  db.Model(post).AddUniqueIndex("index_posts_on_guid", "guid")
  db.Model(post).AddUniqueIndex("index_posts_on_person_id_and_root_guid", "person_id", "root_guid")
  db.Model(post).AddIndex("index_posts_on_person_id", "person_id")
  db.Model(post).AddIndex("index_posts_on_id_and_type_and_created_at", "id", "type", "created_at")
  db.Model(post).AddIndex("index_posts_on_root_guid", "root_guid")
  db.Model(post).AddIndex("index_posts_on_id_and_type", "id", "type")
  db.AutoMigrate(post)

  photo := &Photo{}
  db.Model(photo).AddUniqueIndex("index_photos_on_guid", "guid")
  db.Model(photo).AddIndex("index_photos_on_post_id", "post_id")
  db.Model(photo).AddIndex("index_photos_on_person_id", "person_id")
  db.AutoMigrate(photo)

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
  db.Model(profile).AddUniqueIndex("index_profiles_on_author", "author")
  db.Model(profile).AddIndex("index_profiles_on_full_name_and_searchable", "full_name", "searchable")
  db.Model(profile).AddIndex("index_profiles_on_full_name", "full_name")
  db.AutoMigrate(profile)

  person := &Person{}
  db.Model(person).AddUniqueIndex("index_people_on_author", "author")
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
  db.Model(like).AddUniqueIndex("index_likes_on_shareable_id_and_person_id_and_shareable_type", "shareable_id", "person_id", "shareable_type")
  db.Model(like).AddUniqueIndex("index_likes_on_guid", "guid")
  db.Model(like).AddIndex("likes_person_id_fk", "person_id")
  db.Model(like).AddIndex("index_likes_on_post_id", "shareable_id")
  db.AutoMigrate(like)

  likeSignature := &LikeSignature{}
  db.Model(likeSignature).AddUniqueIndex("index_like_signatures_on_like_id", "like_id")
  db.Model(likeSignature).AddIndex("like_signatures_signature_orders_id_fk", "signature_order_id")
  db.AutoMigrate(likeSignature)

  pod := &Pod{}
  db.Model(pod).AddUniqueIndex("index_pods_on_host", "host")
  db.AutoMigrate(pod)

  shareable := &Shareable{}
  db.Model(shareable).AddUniqueIndex("shareable_and_user_id", "shareable_id", "shareable_type", "user_id")
  db.Model(shareable).AddIndex("index_post_visibilities_on_post_id", "shareable_id")
  db.Model(shareable).AddIndex("index_share_visibilities_on_user_id", "user_id")
  db.Model(shareable).AddIndex("shareable_and_hidden_and_user_id", "shareable_id", "shareable_type", "hidden", "user_id")
  db.AutoMigrate(shareable)

  visibility := &Visibility{}
  db.Model(visibility).AddUniqueIndex("visibilitys_and_person_id", "shareable_id", "shareable_type", "person_id")
  db.Model(visibility).AddIndex("index_visibilitys_on_person_id", "person_id")
  db.Model(visibility).AddIndex("index_visibilitys_on_shareable_id_and_shareable_type", "shareable_id", "shareable_type")
  db.AutoMigrate(visibility)

  aspect := &Aspect{}
  //db.Model(aspect).AddIndex("index_aspects_on_user_id_and_contacts_visible", "user_id", "contacts_visible")
  db.Model(aspect).AddIndex("index_aspects_on_user_id", "user_id")
  db.Model(aspect).AddUniqueIndex("index_aspects_on_user_id_and_name", "user_id", "name")
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

  signatureOrder := &SignatureOrder{}
  db.Model(signatureOrder).AddUniqueIndex("index_signature_orders_on_order", "order")
  db.AutoMigrate(signatureOrder)

  notification := &Notification{}
  db.Model(notification).AddIndex("index_notifications_on_user_id_and_unread", "user_id", "unread")
  db.Model(notification).AddUniqueIndex("index_notifications_on_shareable_type_and_shareable_guid", "shareable_type", "shareable_guid")
  db.AutoMigrate(notification)

  tags := &Tags{}
  db.Model(tags).AddUniqueIndex("index_tags_on_name", "name")
  db.AutoMigrate(tags)

  shareableTagging := &ShareableTagging{}
  db.Model(shareableTagging).AddIndex("index_shareable_tagging_on_id_and_type", "shareable_id", "shareable_type")
  db.Model(shareableTagging).AddUniqueIndex("index_shareable_tagging_on_id_and_type_and_tag_id", "shareable_id", "shareable_type", "tag_id")
  db.AutoMigrate(shareableTagging)

  userTagging := &UserTagging{}
  db.Model(userTagging).AddUniqueIndex("index_user_tagging_on_user_id_and_tag_id", "user_id", "tag_id")
  db.AutoMigrate(userTagging)

  userStream := &UserStream{}
  db.Model(userStream).AddUniqueIndex("index_user_stream_on_user_id_and_name", "user_id", "name")
  db.Model(userStream).AddIndex("index_user_stream_on_user_id", "user_id")
  db.Model(userStream).AddIndex("index_user_stream_on_name", "name")
  db.AutoMigrate(userStream)

  oAuthToken := &OAuthToken{}
  db.Model(oAuthToken).AddUniqueIndex("index_o_auth_token_on_user_id_and_client_id", "user_id", "client_id")
  db.Model(oAuthToken).AddIndex("index_o_auth_token_on_user_id", "user_id")
  db.Model(oAuthToken).AddIndex("index_o_auth_token_on_token", "token")
  db.AutoMigrate(oAuthToken)

  userSetting := &UserSetting{}
  db.Model(userSetting).AddUniqueIndex("index_user_setting_on_user_id_and_key", "user_id", "key")
  db.Model(userSetting).AddIndex("index_user_setting_on_key", "key")
  db.Model(userSetting).AddIndex("index_user_setting_on_key_and_value", "key", "value")
  db.AutoMigrate(userSetting)
}

func InitDB() {
  db, err := OpenDatabase()
  if err != nil {
    panic(err)
  }
  defer db.Close()

  migrationErr := db.Find(&SchemaMigrations{}).Error
  // if no migrations are found its very
  // likely that no schema is loaded at all
  if migrationErr != nil {
    loadSchema(db)
  } else {
    // migrate and print errors to log without aborting
    if err = migrateSchema(db); err != nil {
      revel.AppLog.Error("Something went wrong while migrating", "error", err.Error())
    }
    // incase we had no previous issues we never ran loadSchema
    // but this is still important cause its possible we added
    // a new table or index to it
    loadSchema(db)
  }
}
