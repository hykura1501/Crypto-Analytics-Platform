#!/bin/bash

# Database Restore Script for Crypto Platform
# Usage: ./scripts/restore-db.sh <backup_file>

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if backup file is provided
if [ -z "$1" ]; then
    echo -e "${RED}❌ Error: Backup file is required${NC}"
    echo "Usage: ./scripts/restore-db.sh <backup_file>"
    exit 1
fi

BACKUP_FILE="$1"

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo -e "${RED}❌ Error: Backup file not found: $BACKUP_FILE${NC}"
    exit 1
fi

# Database configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-crypto_db}"

echo -e "${YELLOW}⚠️  WARNING: This will replace the existing database!${NC}"
read -p "Are you sure you want to continue? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Restore cancelled."
    exit 0
fi

echo ""
echo -e "${GREEN}🔄 Starting database restore...${NC}"
echo "Backup file: $BACKUP_FILE"
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo ""

# Check if backup is compressed
TEMP_FILE=""
if [[ "$BACKUP_FILE" == *.gz ]]; then
    echo "Decompressing backup..."
    TEMP_FILE=$(mktemp)
    gunzip -c "$BACKUP_FILE" > "$TEMP_FILE"
    BACKUP_FILE="$TEMP_FILE"
fi

# Check if running in Docker or locally
if docker ps | grep -q crypto_postgres; then
    echo -e "${YELLOW}📦 Detected Docker container, using docker exec...${NC}"
    
    # Restore using docker exec
    cat "$BACKUP_FILE" | docker exec -i crypto_postgres psql -U "$DB_USER" -d postgres
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Database restored successfully!${NC}"
    else
        echo -e "${RED}❌ Restore failed!${NC}"
        [ -n "$TEMP_FILE" ] && rm -f "$TEMP_FILE"
        exit 1
    fi
else
    echo -e "${YELLOW}💻 Using local psql...${NC}"
    
    # Set PGPASSWORD for non-interactive restore
    export PGPASSWORD="$DB_PASSWORD"
    
    # Restore using local psql
    psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d postgres \
        < "$BACKUP_FILE"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Database restored successfully!${NC}"
    else
        echo -e "${RED}❌ Restore failed!${NC}"
        [ -n "$TEMP_FILE" ] && rm -f "$TEMP_FILE"
        unset PGPASSWORD
        exit 1
    fi
    
    unset PGPASSWORD
fi

# Cleanup temp file
[ -n "$TEMP_FILE" ] && rm -f "$TEMP_FILE"

echo ""
echo -e "${GREEN}✨ Restore completed successfully!${NC}"
