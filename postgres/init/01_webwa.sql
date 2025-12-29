DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'webwhatsapp_user') THEN
    CREATE ROLE webwhatsapp_user LOGIN PASSWORD 'WebWhatsappPass123!';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'webwhatsapp_db') THEN
    CREATE DATABASE webwhatsapp_db OWNER webwhatsapp_user;
  END IF;
END $$;
