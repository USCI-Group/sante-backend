# Database Migration Troubleshooting Guide

This guide helpss troubleshoot common Atlas migration issues in the Sante backend project.

## Common Issues and Solutions

### 1. Checksum Mismatch Error

**Error Message:**
```
You have a checksum error in your migration directory.
L2: 20250829020945_init.down.sql was removed
Please check your migration files and run 'atlas migrate hash' to re-hash the contents
Error: checksum mismatch
```

**What This Means:**
- Atlas tracks file checksums to ensure migration integrity
- The error indicates a mismatch between expected and actual file content
- This can happen when files are modified outside of Atlas or there are sync issues

**Solution Steps:**

1. **Navigate to the database directory:**
   ```bash
   cd database
   ```

2. **Regenerate checksums:**
   ```bash
   atlas migrate hash --env local
   ```

3. **Generate new migration (if needed):**
   ```bash
   ./scripts/generate-migration init
   ```

4. **Verify the fix:**
   - Check that no error messages appear
   - Verify new migration files are created in `migrations/` directory
   - Confirm `atlas.sum` file is updated

### 2. Migration File Structure

**Expected Structure:**
```
database/migrations/
├── atlas.sum                    # Checksum file (DO NOT edit manually)
├── 20250829020945_init.up.sql   # Migration files
├── 20250829020945_init.down.sql
├── 20250911024306.up.sql
├── 20250911024306.down.sql
└── [newer migration files...]
```

**Important Notes:**
- Never manually edit `atlas.sum` file
- Always use Atlas commands to manage migrations
- Migration files are named with timestamp prefix

### 3. Atlas Configuration

**Configuration File:** `database/atlas.hcl`
```hcl
data "external_schema" "gorm" {
  program = ["env", "ENCORERUNTIME_NOPANIC=1", "go", "run", "./scripts/atlas-gorm-loader.go"]
}

env "local" {
  src = data.external_schema.gorm.url
  migration {
    dir = "file://migrations"
    format = golang-migrate
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```

### 4. Available Commands

**Generate Migration:**
```bash
./scripts/generate-migration init
```

**Regenerate Checksums:**
```bash
atlas migrate hash --env local
```

**Check Migration Status:**
```bash
atlas migrate status --env local
```

**Apply Migrations:**
```bash
atlas migrate apply --env local
```

### 5. Prevention Tips

1. **Always use Atlas commands** - Don't manually edit migration files
2. **Commit migration files together** - Include both `.up.sql` and `.down.sql` files
3. **Don't modify existing migrations** - Create new ones instead
4. **Keep checksums in sync** - Run `atlas migrate hash` after any file changes
5. **Test migrations locally** - Always test before applying to production

### 6. When to Ask for Help

Contact a senior developer if you encounter:
- Complex schema conflicts
- Data migration issues
- Production migration problems
- Unusual error messages not covered in this guide

### 7. Emergency Recovery

If migrations are completely broken:

1. **Backup your database** (if possible)
2. **Reset migration state:**
   ```bash
   atlas migrate hash --env local
   ```
3. **Regenerate from current schema:**
   ```bash
   ./scripts/generate-migration init
   ```
4. **Test thoroughly** before applying to production

---
