-- 01_create_dbs.sql
-- Creates app role + app database

-- 1) App user (login)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'webwhatsapp_user') THEN
    CREATE ROLE webwhatsapp_user LOGIN PASSWORD 'WebWhatsappPass123!';
  ELSE
    ALTER ROLE webwhatsapp_user WITH LOGIN PASSWORD 'WebWhatsappPass123!';
  END IF;
END
$$;

-- 2) Create DB if missing (match backend: db=webwhatsapp)
SELECT 'CREATE DATABASE webwhatsapp OWNER webwhatsapp_user'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'webwhatsapp')
\gexec

-- 3) Grant
GRANT ALL PRIVILEGES ON DATABASE webwhatsapp TO webwhatsapp_user;
ALTER DATABASE webwhatsapp OWNER TO webwhatsapp_user;
