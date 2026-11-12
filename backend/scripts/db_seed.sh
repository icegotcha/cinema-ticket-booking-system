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
  --file /scripts/js/showtimes.js

echo "Seeding admin..."

mongosh \
  "$MONGODB_URI" \
  --quiet \
  --file /scripts/js/admin.js

echo "Seed Complete"
