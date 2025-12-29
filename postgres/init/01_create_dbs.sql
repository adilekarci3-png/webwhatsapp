-- ./postgres/init/01-init.sql
-- Bu script "ilk init" sırasında (data boşken) çalışır.

DO
$$
BEGIN
  -- Role yoksa oluştur
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'webwhatsapp_user') THEN
    CREATE ROLE webwhatsapp_user LOGIN PASSWORD 'WebWhatsappPass123!';
  END IF;
END
$$;

-- DB yoksa oluştur (NOT: CREATE DATABASE transaction içinde olamaz, bu yüzden DO dışında)
SELECT 'CREATE DATABASE webwhatsapp_db OWNER webwhatsapp_user'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'webwhatsapp_db') \gexec

-- Yetkiler
GRANT ALL PRIVILEGES ON DATABASE webwhatsapp_db TO webwhatsapp_user;

-- public schema standart/temiz yetki
\connect webwhatsapp_db

-- public schema sahibini ayarla
ALTER SCHEMA public OWNER TO webwhatsapp_user;

-- schema üstünde yetki
GRANT ALL ON SCHEMA public TO webwhatsapp_user;

-- default privileges (ileride oluşacak tablolar için)
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO webwhatsapp_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO webwhatsapp_user;

