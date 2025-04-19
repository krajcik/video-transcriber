#!/bin/bash

set -e

VIDEO_DIR="/Users/kraychik/Documents/almedia"
DB_PATH="./db/almedia.db"
SAVETODB_BIN="./cmd/savetodb/savetodb"

for video in "$VIDEO_DIR"/*.mp4; do
  if [ -f "$video" ]; then
    echo "Processing: $video"
    "$SAVETODB_BIN" --video="$video" --db="$DB_PATH"
  fi
done

echo "Batch processing complete. All transcripts saved to $DB_PATH"
