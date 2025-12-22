-- 01_create_dbs.sql

-- 1) Uygulama kullanıcısı (login)
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'webwhatsapp_user') THEN
    CREATE ROLE webwhatsapp_user LOGIN PASSWORD 'WebWhatsappPass123!';
  ELSE
    ALTER ROLE webwhatsapp_user WITH LOGIN PASSWORD 'WebWhatsappPass123!';
  END IF;
END
$$;

-- 2) DB yoksa oluştur
SELECT 'CREATE DATABASE webwhatsapp_db OWNER webwhatsapp_user'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'webwhatsapp_db')
\gexec

-- 3) DB yetkileri
GRANT ALL PRIVILEGES ON DATABASE webwhatsapp_db TO webwhatsapp_user;
