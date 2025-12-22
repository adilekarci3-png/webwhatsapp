-- 02_webwa_schema.sql
\connect webwhatsapp_db;

-- schema yetkileri
GRANT USAGE, CREATE ON SCHEMA public TO webwhatsapp_user;

-- default privileges (ileride yeni tablolar için)
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO webwhatsapp_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO webwhatsapp_user;

-- tabloyu owner olarak webwhatsapp_user ile yarat (en garantisi: SET ROLE)
SET ROLE webwhatsapp_user;

CREATE TABLE IF NOT EXISTS public.messages (
  id              TEXT PRIMARY KEY,
  conversation_id TEXT NOT NULL,
  sender          TEXT NOT NULL,
  body            TEXT NOT NULL,
  created_at_unix BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_ts
ON public.messages (conversation_id, created_at_unix DESC);

RESET ROLE;

-- mevcut tablolar için de yetki ver (SET ROLE kullanmadıysanız bile garanti)
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.messages TO webwhatsapp_user;

-- ====== messages: read receipts + status ======
CREATE INDEX IF NOT EXISTS idx_messages_conv_receiver_read
ON public.messages (conversation_id, receiver, read_at_unix);

ALTER TABLE public.messages ADD COLUMN IF NOT EXISTS receiver TEXT;

ALTER TABLE public.messages
ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'SENT';

ALTER TABLE public.messages
ADD COLUMN IF NOT EXISTS read_at_unix BIGINT;

CREATE INDEX IF NOT EXISTS idx_messages_receiver_read
ON public.messages (receiver, read_at_unix);

CREATE INDEX IF NOT EXISTS idx_messages_conv_receiver_read
ON public.messages (conversation_id, receiver, read_at_unix);