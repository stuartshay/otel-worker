-- Verify Database Schema vs Go Struct Contract
-- This query shows actual database types and sample data

SELECT
    column_name,
    data_type,
    is_nullable,
    CASE
        WHEN column_name = 'id' THEN 'int64'
        WHEN column_name = 'device_id' THEN 'string'
        WHEN column_name = 'tid' THEN 'string'
        WHEN column_name = 'latitude' THEN 'float64'
        WHEN column_name = 'longitude' THEN 'float64'
        WHEN column_name = 'accuracy' THEN 'int'
        WHEN column_name = 'altitude' THEN 'float64'
        WHEN column_name = 'velocity' THEN 'int'
        WHEN column_name = 'battery' THEN 'int'
        WHEN column_name = 'battery_status' THEN 'int'
        WHEN column_name = 'connection_type' THEN 'string'
        WHEN column_name = 'trigger' THEN 'string'
        WHEN column_name = 'timestamp' THEN 'int64 (extracted via EXTRACT(EPOCH))'
        WHEN column_name = 'created_at' THEN 'time.Time'
        WHEN column_name = 'raw_payload' THEN 'not mapped'
    END as go_type,
    CASE
        WHEN data_type = 'integer' AND column_name = 'id' THEN '✅'
        WHEN data_type = 'character varying' AND column_name IN ('device_id', 'tid', 'connection_type', 'trigger') THEN '✅'
        WHEN data_type = 'double precision' AND column_name IN ('latitude', 'longitude', 'altitude') THEN '✅'
        WHEN data_type = 'integer' AND column_name IN ('accuracy', 'velocity', 'battery', 'battery_status') THEN '✅'
        WHEN data_type = 'timestamp with time zone' AND column_name IN ('timestamp', 'created_at') THEN '✅'
        WHEN data_type = 'jsonb' AND column_name = 'raw_payload' THEN '➖'
        ELSE '❌'
    END as match_status
FROM information_schema.columns
WHERE table_name = 'locations'
    AND table_schema = 'public'
ORDER BY ordinal_position;
