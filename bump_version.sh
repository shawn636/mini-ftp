#!/bin/bash

VERSION_FILE="VERSION"

if [ ! -f "$VERSION_FILE" ]; then
  echo "1.0.0" > $VERSION_FILE
fi

current_version=$(cat $VERSION_FILE)
IFS='.' read -r major minor patch <<< "$current_version"

echo "Current version: $current_version"
echo "Choose the version to bump:"
echo "1) Major"
echo "2) Minor"
echo "3) Patch"
read -p "Enter choice [1-3]: " choice

case $choice in
  1)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  2)
    minor=$((minor + 1))
    patch=0
    ;;
  3)
    patch=$((patch + 1))
    ;;
  *)
    echo "Invalid choice!"
    exit 1
    ;;
esac

new_version="$major.$minor.$patch"
echo $new_version > $VERSION_FILE
echo "Version updated to $new_version"