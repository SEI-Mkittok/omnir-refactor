-- Run once on DB creation to enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";   -- for fuzzy text search
CREATE EXTENSION IF NOT EXISTS "btree_gin"; -- for JSONB GIN index support
