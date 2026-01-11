<template>
  <div style="max-width: 1100px; margin: 24px auto; font-family: Arial, sans-serif;">
    <div style="display:flex; align-items:center; gap:12px; margin-bottom:12px;">
      <h2 style="margin: 0;">Web WhatsApp (Demo)</h2>

      <!-- Connection badge -->
      <span :style="connBadgeStyle">
        {{ connLabel }}
      </span>

      <!-- Small meta -->
      <span style="margin-left:auto; font-size:12px; color:#666;">
        Durum:
        <b v-if="typing">yazıyor…</b>
        <b v-else-if="presenceOnline">çevrimiçi</b>
        <b v-else>son görülme {{ lastSeenText }}</b>
      </span>
    </div>

    <!-- Connect Bar -->
    <div style="display:flex; gap:10px; margin-bottom:10px; flex-wrap:wrap; align-items:center;">
      <input v-model="conversationId" placeholder="conversationId / room" style="padding:8px; width:220px;" />
      <input v-model="sender" placeholder="sender / user" style="padding:8px; width:160px;" />
      <input v-model="receiver" placeholder="receiver / to" style="padding:8px; width:160px;" />

      <button @click="connect" :disabled="connected || connecting" style="padding:8px 12px;">
        {{ connecting ? "Bağlanıyor..." : "Bağlan" }}
      </button>
      <button @click="disconnect" :disabled="!connected && !connecting" style="padding:8px 12px;">
        Çık
      </button>

      <label style="display:flex; align-items:center; gap:6px; font-size:12px; color:#444;">
        <input type="checkbox" v-model="autoReconnect" />
        Otomatik tekrar bağlan
      </label>

      <button @click="clearChat" style="padding:8px 12px;">Temizle</button>
    </div>

    <!-- Error banner -->
    <div v-if="errorText"
      style="margin: 8px 0; padding: 10px 12px; border:1px solid #f3c2c2; background:#fff5f5; border-radius:10px; color:#8a1f1f;">
      <div style="display:flex; align-items:center; gap:10px;">
        <b>Hata:</b>
        <span style="flex:1;">{{ errorText }}</span>
        <button @click="errorText = ''" style="padding:6px 10px;">Kapat</button>
      </div>
    </div>

    <!-- Chat Box -->
    <div ref="chatBoxRef"
      style="border:1px solid #ddd; border-radius:10px; height:430px; overflow:auto; background:#f0f2f5; padding:12px;">
      <!-- Loading history -->
      <div v-if="loadingHistory" style="text-align:center; color:#666; padding:14px 0;">
        Geçmiş yükleniyor...
      </div>

      <div v-for="m in messages" :key="m.id" style="display:flex; margin:8px 0;"
        :style="{ justifyContent: m.sender === sender ? 'flex-end' : 'flex-start' }">
        <div :style="bubbleStyle(m)">
          <div style="white-space:pre-wrap;">{{ m.body }}</div>

          <div
            style="display:flex; gap:8px; justify-content:flex-end; align-items:center; margin-top:6px; font-size:11px; color:#666;">
            <span>{{ timeText(m.ts) }}</span>

            <!-- ticks only for my messages -->
            <span v-if="m.sender === sender" :style="{ fontWeight: 700, color: tickColor(m) }">
              {{ tickText(m) }}
            </span>
          </div>
        </div>
      </div>

      <!-- typing indicator bubble -->
      <div v-if="typing" style="display:flex; justify-content:flex-start; margin:8px 0;">
        <div
          style="background:#fff; padding:10px 12px; border-radius:12px; box-shadow:0 1px 2px rgba(0,0,0,0.08); font-size:12px; color:#666;">
          yazıyor…
        </div>
      </div>
    </div>

    <!-- Composer -->
    <div style="display:flex; gap:10px; margin-top:10px; align-items:center;">
      <input v-model="text" placeholder="Mesaj..."
        style="flex:1; padding:12px; border:1px solid #ddd; border-radius:10px;" @keyup.enter="send" @input="onTyping"
        :disabled="!connected" />
      <button @click="send" :disabled="!connected || !text.trim()" style="padding:12px 14px; border-radius:10px;">
        Gönder
      </button>
    </div>

    <!-- Debug / Footer -->
    <div style="margin-top:10px; color:#666; font-size:12px;">
      API Base: {{ apiBase }} | WS: {{ wsBase }} | room(effective): {{ effectiveConversationId }} | me: {{ sender }} |
      to: {{ receiver }}
    </div>

    <!-- Mini log panel -->
    <div style="margin-top:10px; border:1px solid #eee; border-radius:10px; padding:10px; background:#fff;">
      <div style="display:flex; align-items:center; gap:10px;">
        <b style="font-size:12px;">Log</b>
        <span style="font-size:12px; color:#888;">(son 30)</span>
        <button @click="logs = []" style="margin-left:auto; padding:6px 10px;">Log temizle</button>
      </div>
      <div
        style="margin-top:8px; max-height:130px; overflow:auto; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size:11px; color:#333;">
        <div v-for="(l, idx) in logs" :key="idx" style="padding:2px 0; border-bottom:1px dashed #f0f0f0;">
          {{ l }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import { useAuthStore } from "@/stores/auth";
type MessageDTO = {
  id: string;
  conversationId: string;
  sender: string;
  receiver?: string;
  body: string;
  ts: number;
  status?: "SENT" | "ACK" | "READ";
  readAtUnix?: number | null;

  clientMsgId?: string;
  pending?: boolean;
};

type WsEvent =
  | { type: "error"; error: string }
  | { type: "presence.update"; conversationId?: string; sender?: string; payload?: { online: boolean; lastSeenAt?: number } }
  | { type: "typing"; conversationId?: string; sender?: string; payload?: { isTyping: boolean } }
  | { type: "message.read"; conversationId?: string; sender?: string; receiver?: string; payload?: { messageIds: string[]; readAt: number } }
  | { type: "message.ack"; messageId: string; clientMsgId?: string; status?: "ACK" }
  | MessageDTO;

const apiBase = (import.meta.env.VITE_API_BASE as string) || "/api";
const wsBase = (import.meta.env.VITE_WS_URL as string) || "/ws";

const conversationId = ref<string>("general");
const sender = ref<string>("ali");
const receiver = ref<string>("adile");

const text = ref<string>("");
const connected = ref<boolean>(false);
const connecting = ref<boolean>(false);
const loadingHistory = ref<boolean>(false);

const messages = ref<MessageDTO[]>([]);
let ws: WebSocket | null = null;

// presence + typing
const presenceOnline = ref<boolean>(false);
const lastSeenAt = ref<number | null>(null);
const typing = ref<boolean>(false);

// UI helpers
const errorText = ref<string>("");
const logs = ref<string[]>([]);
const autoReconnect = ref<boolean>(true);
let reconnectTimer: number | null = null;

const chatBoxRef = ref<HTMLElement | null>(null);
const auth = useAuthStore(); 

function logLine(s: string) {
  const t = new Date().toLocaleTimeString();
  logs.value.unshift(`[${t}] ${s}`);
  if (logs.value.length > 30) logs.value = logs.value.slice(0, 30);
}

function setError(s: string) {
  errorText.value = s;
  logLine(`ERROR: ${s}`);
}

const connLabel = computed(() => {
  if (connecting.value) return "connecting";
  if (connected.value) return "connected";
  return "disconnected";
});

const connBadgeStyle = computed<Record<string, string>>(() => {
  const base: Record<string, string> = {
    fontSize: "12px",
    padding: "4px 10px",
    borderRadius: "999px",
    border: "1px solid #ddd",
    background: "#fff",
    color: "#333",
  };
  if (connecting.value) {
    base.border = "1px solid #f0d28a";
    base.background = "#fff8e6";
    base.color = "#7a5a00";
  } else if (connected.value) {
    base.border = "1px solid #a7e3b2";
    base.background = "#eefcf1";
    base.color = "#116b2b";
  } else {
    base.border = "1px solid #f3c2c2";
    base.background = "#fff5f5";
    base.color = "#8a1f1f";
  }
  return base;
});

function normalizeUser(u: string) {
  return (u || "").trim().toLowerCase();
}

// DM id: dm:<a>:<b> (sıralı)
function dmConversationId(a: string, b: string): string {
  const x = normalizeUser(a);
  const y = normalizeUser(b);
  if (!x || !y) return "general";
  const [p, q] = [x, y].sort();
  return `dm:${p}:${q}`;
}

// ✅ Tek kaynak: effectiveConversationId
const effectiveConversationId = computed(() => {
  const cid = (conversationId.value || "").trim();

  // Kullanıcı dm: yazdıysa normalize et
  if (cid.startsWith("dm:")) {
    // dm:ali:adile formatını zorla (mevcut değer hatalıysa sender/receiver'dan üret)
    const a = normalizeUser(sender.value);
    const b = normalizeUser(receiver.value);
    return dmConversationId(a, b);
  }

  // custom room adı girilmişse aynen kullan (general hariç)
  if (cid && cid !== "general") return cid;

  // general => eğer receiver varsa DM'e geçmek istiyorsan burada otomatik dm yapabilirsin.
  // Şimdilik: general seçiliyse general kalsın.
  return "general";
});

const wsUrl = computed(() => {
  const token = auth.accessToken || "";

  const q = new URLSearchParams({
    conversationId: effectiveConversationId.value,
    sender: sender.value,
    receiver: receiver.value,
  });

  if (token) q.set("accessToken", token);
  
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
  if (m.pending) return "⏳";
  const st = m.status || "SENT";
  if (st === "SENT") return "✓";
  return "✓✓";
}

function tickColor(m: MessageDTO): string {
  if (m.status === "READ") return "#53bdeb";
  return "#667781";
}

async function loadHistory(): Promise<void> {
  loadingHistory.value = true;
  try {
    const url = `${apiBase}/messages?conversationId=${encodeURIComponent(effectiveConversationId.value)}&limit=50`;

    const token = auth.accessToken; // ✅ store’dan al
    if (!token) {
      throw new Error("Access token yok. Önce login olmalısın.");
    }

    const resp = await fetch(url, {
      method: "GET",
      credentials: "include", // ✅ refresh cookie vs gerekiyorsa
      headers: {
        "Authorization": `Bearer ${token}`, // ✅ kritik
        "Accept": "application/json",
      },
    });

    if (!resp.ok) {
      const t = await resp.text();
      throw new Error(`History ${resp.status}: ${t}`);
    }

    const data = (await resp.json()) as MessageDTO[];
    messages.value = data.slice().reverse();
    logLine(`History loaded (${messages.value.length})`);
    await nextTick();
    scrollToBottom();
  } catch (e: any) {
    setError(e?.message || "History load failed");
  } finally {
    loadingHistory.value = false;
  }
}

function scrollToBottom() {
  const el = chatBoxRef.value;
  if (!el) return;
  el.scrollTop = el.scrollHeight;
}

watch(
  () => messages.value.length,
  async () => {
    await nextTick();
    scrollToBottom();
  }
);

function applyReadEvent(messageIds: string[]): void {
  if (!messageIds?.length) return;
  const set = new Set(messageIds);
  for (const m of messages.value) {
    if (m.sender === sender.value && set.has(m.id)) {
      m.status = "READ";
      m.readAtUnix = Math.floor(Date.now() / 1000);
    }
  }
}

function markUnreadAsRead(): void {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;

  const unreadIds = messages.value
    .filter(m => m.sender !== sender.value)
    .filter(m => !m.readAtUnix && m.status !== "READ")
    .map(m => m.id);

  if (!unreadIds.length) return;

  ws.send(JSON.stringify({
    type: "message.read",
    conversationId: effectiveConversationId.value,
    payload: {
      messageIds: unreadIds,
      readAt: Math.floor(Date.now() / 1000),
    },
  }));
  logLine(`MarkRead sent (${unreadIds.length})`);
}

let typingTimer: number | null = null;

function onTyping(): void {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;

  ws.send(JSON.stringify({
    type: "typing.start",
    conversationId: effectiveConversationId.value,
  }));

  if (typingTimer) window.clearTimeout(typingTimer);
  typingTimer = window.setTimeout(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: "typing.stop",
        conversationId: effectiveConversationId.value,
      }));
    }
    typingTimer = null;
  }, 1500);
}

function scheduleReconnect() {
  if (!autoReconnect.value) return;
  if (reconnectTimer) return;

  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null;
    if (!connected.value && !connecting.value) {
      logLine("Auto reconnect...");
      connect();
    }
  }, 1200);
}

async function connect(): Promise<void> {
  errorText.value = "";
  if (connected.value || connecting.value) return;

  connecting.value = true;
  logLine(`Connecting -> ${wsUrl.value}`);

  await loadHistory();

  try {
    ws = new WebSocket(wsUrl.value);

    ws.onopen = () => {
      connected.value = true;
      connecting.value = false;
      logLine("WS open");
      markUnreadAsRead();
    };

    ws.onmessage = (e: MessageEvent<string>) => {
      let obj: WsEvent | null = null;

      try {
        obj = JSON.parse(e.data) as WsEvent;
      } catch {
        logLine("WS non-JSON frame ignored");
        return;
      }

      if (!obj) return;

      // 1) error
      if ((obj as any).type === "error") {
        setError((obj as any).error || "WS error");
        return;
      }

      // 2) presence
      if ((obj as any).type === "presence.update") {
        const p = (obj as any).payload;
        presenceOnline.value = !!p?.online;
        if (!p?.online && typeof p?.lastSeenAt === "number") lastSeenAt.value = p.lastSeenAt;
        return;
      }

      // 3) typing
      if ((obj as any).type === "typing") {
        const p = (obj as any).payload;
        typing.value = !!p?.isTyping;
        return;
      }

      // 4) read event
      if ((obj as any).type === "message.read") {
        const p = (obj as any).payload;
        if (p?.messageIds?.length) applyReadEvent(p.messageIds);
        return;
      }

      // 5) ack event (pending -> false + ACK)
      if ((obj as any).type === "message.ack") {
        const clientMsgId = (obj as any).clientMsgId as string | undefined;
        const messageId = (obj as any).messageId as string;

        // önce clientMsgId ile bul
        let msg = clientMsgId ? messages.value.find(x => x.clientMsgId === clientMsgId) : undefined;
        // fallback: id ile bul
        if (!msg) msg = messages.value.find(x => x.id === messageId);

        if (msg) {
          msg.id = messageId;     // ✅ gerçek id ile replace
          msg.status = "ACK";
          msg.pending = false;
        }
        return;
      }

      // 6) MessageDTO
      const m = obj as MessageDTO;
      if (!m || !m.id || !m.conversationId) return;

      // ✅ yanlış conversation'dan geleni ignore
      if (m.conversationId !== effectiveConversationId.value) {
        logLine(`Ignored message from other conversation: ${m.conversationId}`);
        return;
      }

      // 6.1) optimistic replace (clientMsgId varsa)
      if (m.clientMsgId) {
        const idx = messages.value.findIndex(x => x.clientMsgId === m.clientMsgId);
        if (idx >= 0) {
          messages.value[idx] = {
            ...messages.value[idx],
            ...m,
            pending: false,
          };

          if (m.sender !== sender.value) {
            setTimeout(() => markUnreadAsRead(), 50);
          }
          return;
        }
      }

      // 6.2) duplicate guard
      if (messages.value.some(x => x.id === m.id)) return;

      // 6.3) normal push
      messages.value.push({ ...m, pending: false });

      if (m.sender !== sender.value) {
        setTimeout(() => markUnreadAsRead(), 50);
      }
    };

    ws.onclose = () => {
      logLine("WS close");
      connected.value = false;
      connecting.value = false;
      ws = null;
      typing.value = false;
      presenceOnline.value = false;
      scheduleReconnect();
    };

    ws.onerror = () => {
      setError("WS error event");
      connected.value = false;
      connecting.value = false;
      scheduleReconnect();
    };
  } catch (e: any) {
    setError(e?.message || "Connect failed");
    connected.value = false;
    connecting.value = false;
    scheduleReconnect();
  }
}

function disconnect(): void {
  if (reconnectTimer) {
    window.clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (ws) ws.close();
  ws = null;
  connected.value = false;
  connecting.value = false;
  typing.value = false;
  presenceOnline.value = false;
  logLine("Disconnected");
}

function makeClientMsgId(): string {
  return `c_${Date.now()}_${Math.random().toString(16).slice(2)}`;
}

function send(): void {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;

  const body = text.value.trim();
  if (!body) return;

  // ✅ DM veya birebirde receiver zorunlu (room’da da isteğe bağlı bırakabilirsin)
  if (effectiveConversationId.value.startsWith("dm:") && !receiver.value.trim()) {
    setError("DM konuşmada receiver boş olamaz.");
    return;
  }

  const now = Math.floor(Date.now() / 1000);
  const clientMsgId = makeClientMsgId();

  // ✅ optimistic mesajı effectiveConversationId ile yaz
  messages.value.push({
    id: clientMsgId,
    clientMsgId,
    conversationId: effectiveConversationId.value, // ✅
    sender: sender.value,
    receiver: receiver.value,
    body,
    ts: now,
    status: "SENT",
    pending: true,
  });

  ws.send(JSON.stringify({
    type: "message.send",
    conversationId: effectiveConversationId.value, // ✅
    receiver: receiver.value,
    body,
    ts: now,
    clientMsgId,
  }));

  text.value = "";
}

function clearChat() {
  messages.value = [];
  logLine("Chat cleared");
}

onBeforeUnmount(() => disconnect());
</script>

