-- 02_webwa_schema.sql
\connect webwhatsapp_db;

-- schema yetkileri
GRANT USAGE, CREATE ON SCHEMA public TO webwhatsapp_user;

-- default privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO webwhatsapp_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO webwhatsapp_user;

-- En garantisi: owner olarak create etmek
SET ROLE webwhatsapp_user;

-- ✅ Tek seferde "final" tablo şeması
CREATE TABLE IF NOT EXISTS public.messages (
  id              TEXT PRIMARY KEY,
  conversation_id TEXT NOT NULL,
  sender          TEXT NOT NULL,
  receiver        TEXT,
  body            TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'SENT',
  created_at_unix BIGINT NOT NULL,
  read_at_unix    BIGINT,
  client_msg_id   TEXT
);

RESET ROLE;

-- mevcut tablolar için de yetki ver
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.messages TO webwhatsapp_user;

-- ✅ Index'ler (kolonlar garanti var)
CREATE INDEX IF NOT EXISTS idx_messages_conversation_ts
  ON public.messages (conversation_id, created_at_unix DESC);

CREATE INDEX IF NOT EXISTS idx_messages_receiver_read
  ON public.messages (receiver, read_at_unix);

CREATE INDEX IF NOT EXISTS idx_messages_conv_receiver_read
  ON public.messages (conversation_id, receiver, read_at_unix);

CREATE INDEX IF NOT EXISTS idx_messages_client_msg_id
  ON public.messages (client_msg_id);
