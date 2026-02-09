#!/bin/bash

# List all database backups
# Usage: ./scripts/list-backups.sh

BACKUP_DIR="${BACKUP_DIR:-./backups}"

if [ ! -d "$BACKUP_DIR" ]; then
    echo "No backups directory found."
    exit 0
fi

echo "📦 Database Backups:"
echo "==================="
echo ""

# List backups with details
ls -lh "$BACKUP_DIR"/*.sql.gz 2>/dev/null | awk '{print $9, "(" $5 ")"}' | while read line; do
    if [ -n "$line" ]; then
        echo "  📄 $line"
    fi
done

# Count backups
COUNT=$(ls -1 "$BACKUP_DIR"/*.sql.gz 2>/dev/null | wc -l)
echo ""
echo "Total: $COUNT backup(s)"
