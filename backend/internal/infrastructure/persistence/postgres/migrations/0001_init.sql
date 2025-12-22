-- messages table (normalized + extensible)

CREATE TABLE IF NOT EXISTS public.messages (
  id                TEXT PRIMARY KEY,
  conversation_id   TEXT NOT NULL,
  sender            TEXT NOT NULL,
  receiver          TEXT,
  body              TEXT NOT NULL,
  status            TEXT NOT NULL DEFAULT 'SENT',
  created_at_unix   BIGINT NOT NULL,
  read_at_unix      BIGINT
);

-- messages listed by conversation, newest first
CREATE INDEX IF NOT EXISTS idx_messages_conversation_ts
  ON public.messages (conversation_id, created_at_unix DESC);

-- unread / read lookup per receiver
CREATE INDEX IF NOT EXISTS idx_messages_receiver_read
  ON public.messages (receiver, read_at_unix);

-- conversation + receiver + read state (WhatsApp-style queries)
CREATE INDEX IF NOT EXISTS idx_messages_conv_receiver_read
  ON public.messages (conversation_id, receiver, read_at_unix);
