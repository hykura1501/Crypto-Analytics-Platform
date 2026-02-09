#!/bin/bash

# Database Backup Script for Crypto Platform
# Usage: ./scripts/backup-db.sh [output_file]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Database configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-crypto_db}"

# Backup directory
BACKUP_DIR="${BACKUP_DIR:-./backups}"
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${1:-${BACKUP_DIR}/crypto_db_backup_${TIMESTAMP}.sql}"
BACKUP_FILE_COMPRESSED="${BACKUP_FILE}.gz"

echo -e "${GREEN}🗄️  Starting database backup...${NC}"
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo ""

# Check if running in Docker or locally
if docker ps | grep -q crypto_postgres; then
    echo -e "${YELLOW}📦 Detected Docker container, using docker exec...${NC}"
    
    # Backup using docker exec
    docker exec crypto_postgres pg_dump \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --clean \
        --if-exists \
        --create \
        --format=plain \
        > "$BACKUP_FILE"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Backup created: $BACKUP_FILE${NC}"
        
        # Compress backup
        echo "Compressing backup..."
        gzip -f "$BACKUP_FILE"
        BACKUP_FILE="${BACKUP_FILE}.gz"
        
        # Get file size
        SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
        echo -e "${GREEN}✅ Compressed backup: $BACKUP_FILE (${SIZE})${NC}"
    else
        echo -e "${RED}❌ Backup failed!${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}💻 Using local pg_dump...${NC}"
    
    # Set PGPASSWORD for non-interactive backup
    export PGPASSWORD="$DB_PASSWORD"
    
    # Backup using local pg_dump
    pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --clean \
        --if-exists \
        --create \
        --format=plain \
        > "$BACKUP_FILE"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Backup created: $BACKUP_FILE${NC}"
        
        # Compress backup
        echo "Compressing backup..."
        gzip -f "$BACKUP_FILE"
        BACKUP_FILE="${BACKUP_FILE}.gz"
        
        # Get file size
        SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
        echo -e "${GREEN}✅ Compressed backup: $BACKUP_FILE (${SIZE})${NC}"
    else
        echo -e "${RED}❌ Backup failed!${NC}"
        exit 1
    fi
    
    unset PGPASSWORD
fi

echo ""
echo -e "${GREEN}✨ Backup completed successfully!${NC}"
echo "📁 Location: $BACKUP_FILE"
echo ""
echo "To restore, run:"
echo "  ./scripts/restore-db.sh $BACKUP_FILE"
