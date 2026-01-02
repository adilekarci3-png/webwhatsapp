<template>
  <div class="auth-page">
    <form class="card" @submit.prevent="onSubmit">
      <h2>Kayıt Ol</h2>

      <label>Ad</label>
      <input v-model.trim="name" required />

      <label>E-posta</label>
      <input v-model.trim="email" type="email" required />

      <label>Şifre</label>
      <input v-model="password" type="password" required minlength="6" />

      <button type="submit" :disabled="loading">
        {{ loading ? "Kaydediliyor..." : "Kayıt Ol" }}
      </button>

      <p class="err" v-if="error">{{ error }}</p>

      <p class="hint">
        Zaten hesabın var mı?
        <router-link to="/login">Giriş Yap</router-link>
      </p>
    </form>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { authApi } from "@/api/auth";

const router = useRouter();
const name = ref("");
const email = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");

async function onSubmit() {
  error.value = "";
  loading.value = true;

  try {
    await authApi.register({ name: name.value, email: email.value, password: password.value });
    await router.push("/login");
  } catch (e) {
    error.value = e?.response?.data?.message || "Kayıt başarısız.";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.auth-page { min-height: 100vh; display:flex; align-items:center; justify-content:center; padding:24px; }
.card { width: 360px; display:flex; flex-direction:column; gap:10px; padding:18px; border:1px solid #ddd; border-radius:12px; background:#fff; }
input { padding:10px; border:1px solid #ccc; border-radius:10px; }
button { padding:10px; border-radius:10px; border:0; cursor:pointer; }
.err { color:#b00020; }
.hint { font-size: 14px; }
</style>
