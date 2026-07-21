#!/usr/bin/env sh
set -e

echo "Seeding movies..."

mongoimport \
  --uri="$MONGODB_URI" \
  --db "$MONGODB_DATABASE" \
  --collection movies \
  --type csv \
  --headerline \
  --file /scripts/source/movies.csv


echo "Seed Complete"