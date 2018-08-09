#!/bin/bash

API=https://api.bintray.com

# NOTE will be set via gitlab protected vars
# BINTRAY_USER=$1
# BINTRAY_API_KEY=$2
# BINTRAY_REPO=$3
PCK_NAME=${UPDATE_CHANNEL}
CURL="curl -u${BINTRAY_USER}:${BINTRAY_API_KEY} -H Content-Type:application/json -H Accept:application/json"

data="{
\"name\": \"${PCK_NAME}\",
\"desc\": \"${PCK_NAME} update channel\",
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
  log=$(mktemp)
  osarch=$(echo ${BIN} |cut -d. -f2)
  upstream="updater.${osarch}.${VERSION}.bin"
  status_code=$(${CURL} --write-out %{http_code} --silent --output ${log} \
    -T ${BIN} -H X-Bintray-Package:${PCK_NAME} \
    -H X-Bintray-Version:${VERSION} ${API}/content/${BINTRAY_REPO}/${upstream})

  if [ $status_code -eq 201 ]; then
    echo "Publishing ${upstream}..."
    echo $(${CURL} -X POST -d "{ \"discard\": \"false\" }" \
      ${API}/content/${BINTRAY_REPO}/${PCK_NAME}/${VERSION}/publish)
  else
    echo -ne "Cannot publish ${upstream}!\n\n" && cat ${log} && exit 1
  fi
  rm ${log}
done
