#!/bin/bash

if [[ "$1" == "mysql" ]]; then
  sed -i 's/db.driver.*/db.driver = "mysql"/' conf/app.conf
  sed -i 's/db.user.*/db.user = "root"/' conf/app.conf
  sed -i 's/db.password.*/db.password = ""/' conf/app.conf
  sed -i 's/db.host.*/db.host = "mysql"/' conf/app.conf
  sed -i 's/db.database.*/db.database = "ganggo"/' conf/app.conf
  sed -i 's#db.dsn.*#db.dsn = "%s:%s@tcp(%s)/%s?parseTime=true"#g' conf/app.conf
  exit 0;
fi

if [[ "$1" == "postgres" ]]; then
  sed -i 's/db.driver.*/db.driver = "postgres"/' conf/app.conf
  sed -i 's/db.user.*/db.user = "postgres"/' conf/app.conf
  sed -i 's/db.password.*/db.password = ""/' conf/app.conf
  sed -i 's/db.host.*/db.host = "postgres"/' conf/app.conf
  sed -i 's/db.database.*/db.database = "ganggo"/' conf/app.conf
  sed -i 's#db.dsn.*#db.dsn = "postgres://%s:%s@%s/%s?sslmode=disable"#g' conf/app.conf
  exit 0;
fi

echo "./$0 [mysql|postgres]";
exit 1;
