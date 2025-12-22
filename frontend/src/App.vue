<template>
  <div style="max-width: 960px; margin: 30px auto; font-family: Arial, sans-serif;">
    <h2>Web WhatsApp (Demo)</h2>

    <div style="display:flex; gap:10px; margin-bottom:10px; flex-wrap:wrap;">
      <input v-model="conversationId" placeholder="conversationId / room" style="padding:8px; width:220px;" />
      <input v-model="sender" placeholder="sender / user" style="padding:8px; width:180px;" />
      <button @click="connect" :disabled="connected" style="padding:8px 12px;">Bağlan</button>
      <button @click="disconnect" :disabled="!connected" style="padding:8px 12px;">Çık</button>
    </div>

    <div style="border:1px solid #ddd; padding:10px; height:380px; overflow:auto; background:#fafafa;">
      <div v-for="m in messages" :key="m.id" style="margin:6px 0;">
        <b>[{{ m.conversationId }}] {{ m.sender }}:</b> {{ m.body }}
        <span style="color:#666; font-size:12px;">
          ({{ new Date(m.ts * 1000).toLocaleTimeString() }})
        </span>
      </div>
    </div>

    <div style="display:flex; gap:10px; margin-top:10px;">
      <input
        v-model="text"
        placeholder="Mesaj..."
        style="flex:1; padding:10px;"
        @keyup.enter="send"
      />
      <button @click="send" :disabled="!connected || !text.trim()" style="padding:10px 14px;">Gönder</button>
    </div>

    <div style="margin-top:10px; color:#666; font-size:12px;">
      API Base: {{ apiBase }} | WS: {{ wsBase }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from "vue";

type MessageDTO = {
  id: string;
  conversationId: string;
  sender: string;
  body: string;
  ts: number;
};

const apiBase = (import.meta.env.VITE_API_BASE as string) || "/api";
const wsBase = (import.meta.env.VITE_WS_URL as string) || "/ws";

const conversationId = ref<string>("general");
const sender = ref<string>("ali");

const text = ref<string>("");
const connected = ref<boolean>(false);
const messages = ref<MessageDTO[]>([]);

let ws: WebSocket | null = null;

const wsUrl = computed(() => {
  const q = new URLSearchParams({
    conversationId: conversationId.value,
    sender: sender.value,
  });

  // wsBase "/ws" ise aynı origin üzerinden ws'e çeviriyoruz
  if (wsBase.startsWith("ws://") || wsBase.startsWith("wss://")) {
    return `${wsBase}?${q.toString()}`;
  }
  const originWs = location.origin.replace("http://", "ws://").replace("https://", "wss://");
  return `${originWs}${wsBase}?${q.toString()}`;
});

async function loadHistory(): Promise<void> {
  const url = `${apiBase}/messages?conversationId=${encodeURIComponent(conversationId.value)}&limit=50`;
  const resp = await fetch(url);
  const data = (await resp.json()) as MessageDTO[];
  // backend DESC döndürüyor; UI'da ASC göstermek için reverse
  messages.value = data.slice().reverse();
}

async function connect(): Promise<void> {
  await loadHistory();

  ws = new WebSocket(wsUrl.value);

  ws.onopen = () => {
    connected.value = true;
  };

  ws.onmessage = (e: MessageEvent<string>) => {
    try {
      const obj = JSON.parse(e.data) as any;
      // error envelope geldiyse
      if (obj?.type === "error") return;

      const m = obj as MessageDTO;
      if (m && m.id) {
        messages.value.push(m);
      }
    } catch {
      // ignore
    }
  };

  ws.onclose = () => {
    connected.value = false;
    ws = null;
  };

  ws.onerror = () => {
    connected.value = false;
  };
}

function disconnect(): void {
  if (ws) ws.close();
  ws = null;
  connected.value = false;
}

function send(): void {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;
  const body = text.value.trim();
  if (!body) return;
  ws.send(body);
  text.value = "";
}

onBeforeUnmount(() => disconnect());
</script>
