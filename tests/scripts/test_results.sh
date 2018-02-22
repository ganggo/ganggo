#!/bin/bash

find test-results -name '*.failed.html' |while read file; do
  echo -e "\033[0;31m$file\033[0m";
  cat $file |tr '\n' ';' |grep -oP '(?<=<code class="bash">).*?(?=</code>)' \
    | sed -r 's/\);\s*/\)\n/g';
  echo -e "\n\n";
done
