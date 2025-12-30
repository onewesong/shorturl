<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { apiListLinks, apiLogout } from "../api";
import { resetAuthCache } from "../router";

const router = useRouter();
const loading = ref(false);
const error = ref("");
const links = ref([]);

const origin = computed(() => window.location.origin);
const shortUrl = (code) => `${origin.value}/${code}`;

async function load() {
  loading.value = true;
  error.value = "";
  try {
    links.value = await apiListLinks();
  } catch (e) {
    error.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

async function logout() {
  await apiLogout();
  resetAuthCache();
  await router.replace("/login");
}

async function copy(text) {
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    // ignore
  }
}

onMounted(load);
</script>

<template>
  <div class="mx-auto max-w-5xl px-4 py-6">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">短链管理</h1>
      <div class="flex gap-2">
        <button
          class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm hover:bg-slate-50"
          type="button"
          @click="logout"
        >
          退出
        </button>
        <button
          class="rounded-lg bg-slate-900 px-3 py-2 text-sm text-white hover:bg-slate-800"
          type="button"
          @click="router.push('/links/new')"
        >
          新建
        </button>
      </div>
    </div>

    <div v-if="error" class="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>

    <div class="mt-4 overflow-hidden rounded-xl border border-slate-200 bg-white">
      <div v-if="loading" class="p-4 text-sm text-slate-600">加载中...</div>
      <table v-else class="w-full table-auto text-sm">
        <thead class="bg-slate-50 text-left text-slate-600">
          <tr>
            <th class="px-4 py-3">短码</th>
            <th class="px-4 py-3">目标URL</th>
            <th class="px-4 py-3">状态</th>
            <th class="px-4 py-3">点击</th>
            <th class="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <tr v-for="l in links" :key="l.id" class="align-top">
            <td class="px-4 py-3 font-mono">
              <a class="text-slate-900 underline" :href="`/${l.code}`" target="_blank" rel="noreferrer">{{ l.code }}</a>
              <div class="mt-1 flex gap-2">
                <button
                  type="button"
                  class="rounded-lg border border-slate-200 bg-white px-2 py-1 text-xs hover:bg-slate-50"
                  @click="copy(shortUrl(l.code))"
                >
                  复制
                </button>
              </div>
            </td>
            <td class="px-4 py-3 break-all text-slate-700">{{ l.target_url }}</td>
            <td class="px-4 py-3">
              <span
                v-if="l.enabled"
                class="inline-flex rounded-full bg-green-50 px-2 py-1 text-xs text-green-700"
                >启用</span
              >
              <span v-else class="inline-flex rounded-full bg-slate-100 px-2 py-1 text-xs text-slate-700">禁用</span>
            </td>
            <td class="px-4 py-3 text-slate-700">{{ l.click_count }}</td>
            <td class="px-4 py-3 text-right">
              <button
                type="button"
                class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs hover:bg-slate-50"
                @click="router.push(`/links/${l.id}/edit`)"
              >
                编辑
              </button>
            </td>
          </tr>
          <tr v-if="links.length === 0">
            <td class="px-4 py-6 text-center text-sm text-slate-500" colspan="5">暂无短链，点击右上角新建。</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

