#!/bin/bash

########################################################################
# Package the binaries built on Travis-CI as an AppImage
# By Simon Peter 2016
# For more information, see http://appimage.org/
########################################################################

set -e

export ARCH="$(arch)"

APP=cmake-tools
LOWERAPP=${APP,,}

mkdir -p "$HOME/$APP/$APP.AppDir/usr/"

BUILD_PATH="$(pwd)"

cd "$HOME/$APP/"

wget -q https://github.com/probonopd/AppImages/raw/master/functions.sh -O ./functions.sh
. ./functions.sh

cd $APP.AppDir

cp "${BUILD_PATH}/${LOWERAPP}" ./usr/bin
cp -r "${BUILD_PATH}/cmake" ./usr/bin
cp "${BUILD_PATH}/.clang-format" ./usr/bin
cp "${BUILD_PATH}/.clang-tidy" ./usr/bin

########################################################################
# Copy desktop and icon file to AppDir for AppRun to pick them up
########################################################################

cp "${BUILD_PATH}/appimage/${LOWERAPP}.desktop" .
cp "${BUILD_PATH}/appimage/${LOWERAPP}.png" .

########################################################################
# Other appliaction-specific finishing touches
########################################################################

# Bundle Python and all the plugins needed

cd ..

########################################################################
# Copy in the dependencies that cannot be assumed to be available
# on all target systems
########################################################################

copy_deps

########################################################################
# desktopintegration asks the user on first run to install a menu item
########################################################################

get_desktopintegration $LOWERAPP

########################################################################
# Determine the version of the app; also include needed glibc version
########################################################################

VERSION="0.01"

########################################################################
# Patch away absolute paths; it would be nice if they were relative
########################################################################

find usr/ -type f -exec sed -i -e 's|/usr|././|g' {} \;
find usr/ -type f -exec sed -i -e 's@././/bin/env@/usr/bin/env@g' {} \;

########################################################################
# AppDir complete
# Now packaging it as an AppImage
########################################################################

cd .. # Go out of AppImage

mkdir -p ../out/
generate_type2_appimage

########################################################################
# Upload the AppDir
########################################################################

transfer ../out/*
echo "AppImage has been uploaded to the URL above; use something like GitHub Releases for permanent storage"
