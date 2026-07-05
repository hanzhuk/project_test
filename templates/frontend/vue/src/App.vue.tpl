<script setup lang="ts">
import { ref, onMounted } from 'vue';

const status = ref('检查中...');

onMounted(async () => {
  try {
    const res = await fetch('http://localhost:8080/health');
    const data = await res.json();
    if (data.status === 'ok') {
      status.value = '后端服务正常 √';
    } else {
      status.value = '后端异常';
    }
  } catch (err) {
    status.value = '无法连接后端';
  }
});
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex items-center justify-center p-6">
    <div class="bg-white p-8 rounded-xl shadow-md max-w-md w-full border border-slate-100">
      <h1 class="text-2xl font-bold text-slate-800 mb-2">{{.ProjectName}}</h1>
      <p class="text-slate-500 text-sm mb-6">基于 Go + Echo + Ent + Vue 3 的全栈应用</p>

      <div class="p-4 bg-slate-50 rounded-lg text-sm text-slate-700 space-y-2">
        <div>服务状态: <span class="font-medium text-emerald-600">{{ status }}</span></div>
        <div>前端框架: <span class="font-medium">Vue 3 + Pinia + Vite</span></div>
      </div>
    </div>
  </div>
</template>
