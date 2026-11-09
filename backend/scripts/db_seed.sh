#!/usr/bin/env sh
set -e

echo "Seeding movies..."

mongoimport \
  --uri="$MONGODB_URI" \
  --db "$MONGODB_DATABASE" \
  --collection movies \
  --drop \
  --type csv \
  --headerline \
  --file /scripts/source/movies.csv

echo "Seeding showtimes..."

mongosh \
  "$MONGODB_URI" \
  --quiet \
  --file /scripts/source/showtimes.js

echo "Seed Complete"
