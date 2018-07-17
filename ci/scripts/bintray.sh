#!/bin/bash

API=https://api.bintray.com

# NOTE will be set via gitlab protected vars
# BINTRAY_USER=$1
# BINTRAY_API_KEY=$2
# BINTRAY_REPO=$3
PCK_NAME=${UPDATE_CHANNEL}
PCK_VERSION=$(git describe --abbrev=0 --tags)
CURL="curl -u${BINTRAY_USER}:${BINTRAY_API_KEY} -H Content-Type:application/json -H Accept:application/json"

data="{
\"name\": \"${PCK_NAME}\",
\"desc\": \"bump to v${PCK_VERSION}\",
\"vcs_url\": \"${CI_PROJECT_URL}\",
\"licenses\": [\"GPL-3.0\"],
\"issue_tracker_url\": \"${CI_PROJECT_URL}/issues\",
\"website_url\": \"${CI_PROJECT_URL}/wikis/home\",
\"desc_url\": \"${CI_PROJECT_URL}\",
\"labels\": [\"ganggo\", \"gitlab\", \"deploy\"],
\"public_download_numbers\": true,
\"public_stats\": true
}"

echo "Creating package ${PCK_NAME}..."
echo $(${CURL} -X POST -d "${data}" ${API}/packages/${BINTRAY_REPO})

for BIN in $(ls updater.*.bin); do
  status_code=$(${CURL} --write-out %{http_code} --silent --output /dev/null \
    -T ${BIN} -H X-Bintray-Package:${PCK_NAME} \
    -H X-Bintray-Version:${PCK_VERSION} ${API}/content/${BINTRAY_REPO}/${BIN})

  if [ $status_code -eq 201 ]; then
    echo "Publishing ${BIN}..."
    echo $(${CURL} -X POST -d "{ \"discard\": \"false\" }" \
      ${API}/content/${BINTRAY_REPO}/${PCK_NAME}/${PCK_VERSION}/publish)
  else
    echo "Cannot publish ${BIN}!"
  fi
done
