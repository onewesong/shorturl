<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { apiCreateLink, apiGetLink, apiUpdateLink } from "../api";

const router = useRouter();
const route = useRoute();

const id = computed(() => route.params.id);
const isEdit = computed(() => typeof id.value === "string" && id.value.length > 0);

const loading = ref(false);
const error = ref("");

const code = ref("");
const targetUrl = ref("");
const enabled = ref(true);
const customCode = ref("");

async function load() {
  if (!isEdit.value) return;
  loading.value = true;
  error.value = "";
  try {
    const link = await apiGetLink(id.value);
    code.value = link.code;
    targetUrl.value = link.target_url;
    enabled.value = !!link.enabled;
  } catch (e) {
    error.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

async function submit() {
  loading.value = true;
  error.value = "";
  try {
    if (isEdit.value) {
      await apiUpdateLink(id.value, { target_url: targetUrl.value.trim(), enabled: enabled.value });
    } else {
      await apiCreateLink({ target_url: targetUrl.value.trim(), code: customCode.value.trim() || undefined });
    }
    await router.replace("/links");
  } catch (e) {
    error.value = e?.message || "保存失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="mx-auto max-w-5xl px-4 py-6">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">{{ isEdit ? "编辑短链" : "新建短链" }}</h1>
      <button class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm hover:bg-slate-50" @click="router.push('/links')">
        返回
      </button>
    </div>

    <div v-if="error" class="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>

    <div class="mt-4 rounded-xl border border-slate-200 bg-white p-6">
      <form class="space-y-4" @submit.prevent="submit">
        <div v-if="isEdit" class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="block text-sm font-medium">短码</label>
            <input
              :value="code"
              disabled
              class="mt-1 w-full rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 font-mono text-slate-700"
            />
          </div>
          <div class="flex items-end">
            <label class="inline-flex items-center gap-2">
              <input v-model="enabled" type="checkbox" class="h-4 w-4" />
              <span class="text-sm">启用</span>
            </label>
          </div>
        </div>

        <div v-else>
          <label class="block text-sm font-medium">短码（可选）</label>
          <input
            v-model="customCode"
            placeholder="留空自动生成"
            class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 font-mono focus:outline-none focus:ring-2 focus:ring-slate-400"
          />
          <p class="mt-1 text-xs text-slate-500">允许字母/数字/ - / _</p>
        </div>

        <div>
          <label class="block text-sm font-medium">目标URL</label>
          <input
            v-model="targetUrl"
            placeholder="https://example.com/..."
            class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-slate-400"
            required
          />
          <p class="mt-1 text-xs text-slate-500">仅支持 http/https。</p>
        </div>

        <button
          class="rounded-lg bg-slate-900 px-4 py-2 text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          type="submit"
          :disabled="loading"
        >
          {{ loading ? "保存中..." : "保存" }}
        </button>
      </form>
    </div>
  </div>
</template>

