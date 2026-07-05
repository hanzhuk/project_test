<script lang="ts">
  import { onMount } from 'svelte';

  let status = $state('检查中...');

  onMount(async () => {
    try {
      const res = await fetch('http://localhost:8080/health');
      const data = await res.json();
      if (data.status === 'ok') {
        status = '后端服务正常 √';
      } else {
        status = '后端异常';
      }
    } catch (err) {
      status = '无法连接后端';
    }
  });
</script>

<main class="min-h-screen bg-slate-50 flex items-center justify-center p-6">
  <div class="bg-white p-8 rounded-xl shadow-md max-w-md w-full border border-slate-100">
    <h1 class="text-2xl font-bold text-slate-800 mb-2">{{.ProjectName}}</h1>
    <p class="text-slate-500 text-sm mb-6">基于 Go + Echo + Ent + Svelte 5 的全栈应用</p>

    <div class="p-4 bg-slate-50 rounded-lg text-sm text-slate-700 space-y-2">
      <div>服务状态: <span class="font-medium text-emerald-600">{status}</span></div>
      <div>前端框架: <span class="font-medium">Svelte 5 + Vite</span></div>
    </div>
  </div>
</main>
