#!/bin/bash
set -e

# TODO - check lein is installed
lein uberjar

rm -rf release
mkdir release
pushd release

# Get bitbar from remote
# TODO - allow flag for specifying version and sane default
wget https://github.com/matryer/bitbar/releases/download/v1.9.2/BitBarDistro-v1.9.2.zip
wget -O bundler.sh https://raw.githubusercontent.com/matryer/bitbar/v1.9.2/Scripts/bitbar-bundler
chmod +x bundler.sh

rm -rf BitBarDistro.app
unzip BitBarDistro-v1.9.2.zip

cp -R BitBarDistro.app Prowler.app

# Bundle the app
./bundler.sh Prowler.app ../bitbar.1m.sh

mkdir -p Prowler.app/Contents/SharedSupport/bin
cp  ../target/prowler-0.1.0-SNAPSHOT-standalone.jar Prowler.app/Contents/SharedSupport/bin/
chmod +x Prowler.app/Contents/SharedSupport/bin/prowler-0.1.0-SNAPSHOT-standalone.jar
popd