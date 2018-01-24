#!/bin/bash

docker run --name=d1 -e DATABASE=d1 \
  -e PORT=3000 -p 3000:3000 --net=host -d ganggo/testing_diaspora:v1.0.0

docker run --name=d2 -e DATABASE=d2 \
  -e PORT=3001 -p 3001:3001 --net=host -d ganggo/testing_diaspora:v1.0.0

echo -e "\033[0;32mServer started complete\033[0m"

# Wait till they finished loading
for port in 3000 3001; do
  i=0
  while [[ "$(curl http://localhost:$port >/dev/null 2>&1; echo $?)" -ne "0" ]]; do
    echo "Waiting for localhost:$port"
    sleep 2
    if [ $i -gt 320 ]; then
      echo ".. timeout!"
      exit 1
    else
      ((i++))
    fi
  done
done

echo -e "\033[0;32mServer setup complete\033[0m"

# XXX hack otherwise first hcard test will run into a timeout
curl http://localhost:3000/hcard/users/2d4fa7e0e5380135fa593c970e8692d1 >/dev/null 2>&1
curl http://localhost:3001/hcard/users/2d4fa7e0e5380135fa593c970e8692d2 >/dev/null 2>&1

echo -e "\033[0;32mServer configuration complete\033[0m"
