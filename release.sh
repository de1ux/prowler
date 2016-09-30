#!/bin/bash
set -e

lein uberjar

pushd bitrelease

rm -rf BitBarDistro.app
unzip BitBarDistro-v1.9.1.zip

cp -R BitBarDistro.app Prowler.app
./bundler.sh Prowler.app bitbar.1m.sh
mkdir -p Prowler.app/Contents/SharedSupport/bin
cp  ../target/prowler-0.1.0-SNAPSHOT-standalone.jar Prowler.app/Contents/SharedSupport/bin/
chmod +x Prowler.app/Contents/SharedSupport/bin/prowler-0.1.0-SNAPSHOT-standalone.jar
popd

cp -R bitrelease/Prowler.app .