<script setup>
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { apiLogin } from "../api";
import { resetAuthCache } from "../router";

const router = useRouter();
const route = useRoute();
const username = ref("admin");
const password = ref("");
const error = ref("");
const loading = ref(false);

async function submit() {
  error.value = "";
  loading.value = true;
  try {
    await apiLogin(username.value.trim(), password.value);
    resetAuthCache();
    const next = typeof route.query.next === "string" ? route.query.next : "/links";
    await router.replace(next);
  } catch (e) {
    error.value = e?.message || "登录失败";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="mx-auto max-w-5xl px-4 py-6">
    <div class="mx-auto mt-20 max-w-md rounded-xl bg-white p-6 shadow">
      <h1 class="text-xl font-semibold">短链后台登录</h1>
      <p class="mt-2 text-sm text-slate-600">使用管理员账号登录后管理短链。</p>

      <div v-if="error" class="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>

      <form class="mt-6 space-y-4" @submit.prevent="submit">
        <div>
          <label class="block text-sm font-medium">用户名</label>
          <input
            v-model="username"
            class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-slate-400"
            autocomplete="username"
            required
          />
        </div>
        <div>
          <label class="block text-sm font-medium">密码</label>
          <input
            v-model="password"
            type="password"
            class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-slate-400"
            autocomplete="current-password"
            required
          />
        </div>
        <button
          class="w-full rounded-lg bg-slate-900 px-4 py-2 text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          type="submit"
          :disabled="loading"
        >
          {{ loading ? "登录中..." : "登录" }}
        </button>
      </form>
    </div>
  </div>
</template>

