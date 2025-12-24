<template>
  <div style="max-width: 1100px; margin: 24px auto; font-family: Arial, sans-serif;">
    <h2 style="margin: 0 0 12px;">Web WhatsApp (Demo)</h2>

    <!-- Connect Bar -->
    <div style="display:flex; gap:10px; margin-bottom:10px; flex-wrap:wrap; align-items:center;">
      <input v-model="conversationId" placeholder="conversationId / room" style="padding:8px; width:220px;" />
      <input v-model="sender" placeholder="sender / user" style="padding:8px; width:160px;" />
      <input v-model="receiver" placeholder="receiver / to" style="padding:8px; width:160px;" />
      <button @click="connect" :disabled="connected" style="padding:8px 12px;">Bağlan</button>
      <button @click="disconnect" :disabled="!connected" style="padding:8px 12px;">Çık</button>

      <span style="margin-left:auto; font-size:12px; color:#666;">
        Durum:
        <b v-if="typing">yazıyor…</b>
        <b v-else-if="presenceOnline">çevrimiçi</b>
        <b v-else>son görülme {{ lastSeenText }}</b>
      </span>
    </div>

    <!-- Chat Box -->
    <div style="border:1px solid #ddd; border-radius:10px; height:430px; overflow:auto; background:#f0f2f5; padding:12px;">
      <div v-for="m in messages" :key="m.id" style="display:flex; margin:8px 0;"
           :style="{ justifyContent: m.sender === sender ? 'flex-end' : 'flex-start' }">
        <div
          :style="bubbleStyle(m)"
        >
          <div style="white-space:pre-wrap;">{{ m.body }}</div>

          <div style="display:flex; gap:8px; justify-content:flex-end; align-items:center; margin-top:6px; font-size:11px; color:#666;">
            <span>{{ timeText(m.ts) }}</span>

            <!-- ticks only for my messages -->
            <span v-if="m.sender === sender" :style="{ fontWeight: 700, color: tickColor(m) }">
              {{ tickText(m) }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Composer -->
    <div style="display:flex; gap:10px; margin-top:10px; align-items:center;">
      <input
        v-model="text"
        placeholder="Mesaj..."
        style="flex:1; padding:12px; border:1px solid #ddd; border-radius:10px;"
        @keyup.enter="send"
        @input="onTyping"
      />
      <button @click="send" :disabled="!connected || !text.trim()" style="padding:12px 14px; border-radius:10px;">
        Gönder
      </button>
    </div>

    <div style="margin-top:10px; color:#666; font-size:12px;">
      API Base: {{ apiBase }} | WS: {{ wsBase }} | room: {{ conversationId }} | me: {{ sender }} | to: {{ receiver }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from "vue";

type MessageDTO = {
  id: string;
  conversationId: string;
  sender: string;
  receiver?: string;
  body: string;
  ts: number;
  status?: "SENT" | "ACK" | "READ";
  readAtUnix?: number | null;
};

// generic WS envelope
type WsEvent =
  | { type: "error"; error: string }
  | { type: "presence.update"; conversationId?: string; sender?: string; payload?: { online: boolean; lastSeenAt?: number } }
  | { type: "typing"; conversationId?: string; sender?: string; payload?: { isTyping: boolean } }
  | { type: "message.read"; conversationId?: string; sender?: string; receiver?: string; payload?: { messageIds: string[]; readAt: number } }
  | { type: "message.ack"; messageId: string; status?: "ACK" }
  | MessageDTO;

const apiBase = (import.meta.env.VITE_API_BASE as string) || "/api";
const wsBase = (import.meta.env.VITE_WS_URL as string) || "/ws";

const conversationId = ref<string>("general");
const sender = ref<string>("ali");
const receiver = ref<string>("adile"); // ✅ 1-1 okundu için gerekli

const text = ref<string>("");
const connected = ref<boolean>(false);
const messages = ref<MessageDTO[]>([]);

let ws: WebSocket | null = null;

// presence + typing state
const presenceOnline = ref<boolean>(false);
const lastSeenAt = ref<number | null>(null);
const typing = ref<boolean>(false);

let typingTimer: number | null = null;

const wsUrl = computed(() => {
  const q = new URLSearchParams({
    conversationId: conversationId.value,
    sender: sender.value,
    receiver: receiver.value, // ✅
  });

  if (wsBase.startsWith("ws://") || wsBase.startsWith("wss://")) {
    return `${wsBase}?${q.toString()}`;
  }
  const originWs = location.origin.replace("http://", "ws://").replace("https://", "wss://");
  return `${originWs}${wsBase}?${q.toString()}`;
});

function timeText(ts: number): string {
  return new Date(ts * 1000).toLocaleTimeString();
}

const lastSeenText = computed(() => {
  if (!lastSeenAt.value) return "-";
  return new Date(lastSeenAt.value * 1000).toLocaleTimeString();
});

function bubbleStyle(m: MessageDTO): Record<string, string> {
  const mine = m.sender === sender.value;
  return {
    maxWidth: "70%",
    background: mine ? "#d9fdd3" : "#fff",
    padding: "10px 12px",
    borderRadius: "12px",
    boxShadow: "0 1px 2px rgba(0,0,0,0.08)",
  };
}

function tickText(m: MessageDTO): string {
  // minimal: SENT -> ✓, ACK/READ -> ✓✓
  const st = m.status || "SENT";
  if (st === "SENT") return "✓";
  return "✓✓";
}

function tickColor(m: MessageDTO): string {
  // READ -> whatsapp blue
  if (m.status === "READ") return "#53bdeb";
  return "#667781";
}

async function loadHistory(): Promise<void> {
  const url = `${apiBase}/messages?conversationId=${encodeURIComponent(conversationId.value)}&limit=50`;
  const resp = await fetch(url);
  const data = (await resp.json()) as MessageDTO[];
  messages.value = data.slice().reverse();
}

function applyReadEvent(messageIds: string[]): void {
  if (!messageIds?.length) return;
  const set = new Set(messageIds);
  for (const m of messages.value) {
    // karşı taraf benim mesajlarımı okuduysa => benim mesajlarım READ
    if (m.sender === sender.value && set.has(m.id)) {
      m.status = "READ";
      m.readAtUnix = Math.floor(Date.now() / 1000);
    }
  }
}

// sohbet açılınca / yeni mesaj gelince, karşıdan gelen unread mesajları okundu yap
function markUnreadAsRead(): void {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;

  const unreadIds = messages.value
    .filter(m => m.sender !== sender.value)        // karşıdan gelen
    .filter(m => !m.readAtUnix && m.status !== "READ")
    .map(m => m.id);

  if (!unreadIds.length) return;

  ws.send(JSON.stringify({
    type: "message.read",
    conversationId: conversationId.value,
    payload: {
      messageIds: unreadIds,
      readAt: Math.floor(Date.now() / 1000),
    },
  }));
}

function onTyping(): void {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;

  // start typing (throttle)
  ws.send(JSON.stringify({ type: "typing.start", conversationId: conversationId.value }));

  if (typingTimer) window.clearTimeout(typingTimer);
  typingTimer = window.setTimeout(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: "typing.stop", conversationId: conversationId.value }));
    }
    typingTimer = null;
  }, 1500);
}

async function connect(): Promise<void> {
  await loadHistory();

  ws = new WebSocket(wsUrl.value);

  ws.onopen = () => {
    connected.value = true;
    // sohbet açılınca mevcut unread’ları okundu yap
    markUnreadAsRead();
  };

  ws.onmessage = (e: MessageEvent<string>) => {
    let obj: WsEvent | null = null;

    try {
      obj = JSON.parse(e.data) as WsEvent;
    } catch {
      // eski plain-text akış gelirse ignore (sende artık JSON DTO bekliyoruz)
      return;
    }

    if (!obj) return;

    if ((obj as any).type === "error") return;

    // presence
    if ((obj as any).type === "presence.update") {
      const p = (obj as any).payload;
      if (p?.online === true) {
        presenceOnline.value = true;
      } else {
        presenceOnline.value = false;
        if (typeof p?.lastSeenAt === "number") lastSeenAt.value = p.lastSeenAt;
      }
      return;
    }

    // typing
    if ((obj as any).type === "typing") {
      const p = (obj as any).payload;
      typing.value = !!p?.isTyping;
      return;
    }

    // read event
    if ((obj as any).type === "message.read") {
      const p = (obj as any).payload;
      if (p?.messageIds?.length) applyReadEvent(p.messageIds);
      return;
    }

    // ack (opsiyonel)
    if ((obj as any).type === "message.ack") {
      const id = (obj as any).messageId;
      const msg = messages.value.find(x => x.id === id);
      if (msg) msg.status = "ACK";
      return;
    }

    // message dto
    const m = obj as MessageDTO;
    if (m && m.id && m.conversationId) {
      messages.value.push(m);

      // yeni mesaj karşıdan geldiyse otomatik okundu yap
      if (m.sender !== sender.value) {
        // küçük gecikme: UI render sonrası
        setTimeout(() => markUnreadAsRead(), 50);
      }
    }
  };

  ws.onclose = () => {
    connected.value = false;
    ws = null;
    typing.value = false;
    presenceOnline.value = false;
  };

  ws.onerror = () => {
    connected.value = false;
  };
}

function disconnect(): void {
  if (ws) ws.close();
  ws = null;
  connected.value = false;
  typing.value = false;
  presenceOnline.value = false;
}

function send(): void {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;
  const body = text.value.trim();
  if (!body) return;

  ws.send(JSON.stringify({
    type: "message.send",
    conversationId: conversationId.value,
    sender: sender.value,
    receiver: receiver.value,
    body,
    ts: Math.floor(Date.now() / 1000),
  }));

  text.value = "";
}

onBeforeUnmount(() => disconnect());
</script>
