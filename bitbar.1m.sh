#!/bin/bash

out=$(java -jar /Applications/Prowler.app/Contents/SharedSupport/bin/prowler-0.1.0-SNAPSHOT-standalone.jar)
export IFS=";"
for word in $out; do
    echo $word
done