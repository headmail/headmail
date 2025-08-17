<template>
  <div class="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 space-y-4">
    <div class="flex items-center gap-3">
      <div class="flex items-center gap-2">
        <label class="text-sm text-gray-600">기간</label>
        <select v-model="preset" @change="applyPreset" class="px-2 py-1 border rounded bg-white">
          <option value="24h">지난 24시간</option>
          <option value="7d">지난 7일</option>
          <option value="30d">지난 30일</option>
          <option value="custom">직접 선택</option>
        </select>
      </div>

      <div v-if="preset === 'custom'" class="flex items-center gap-2">
        <input type="datetime-local" v-model="fromLocal" @change="onCustomDateChange" class="px-2 py-1 border rounded" />
        <span class="text-sm text-gray-500">~</span>
        <input type="datetime-local" v-model="toLocal" @change="onCustomDateChange" class="px-2 py-1 border rounded" />
      </div>

      <div class="flex items-center gap-2">
        <label class="text-sm text-gray-600">Granularity</label>
        <select v-model="granularity" @change="reload" class="px-2 py-1 border rounded bg-white">
          <option value="hour">Hour</option>
          <option value="day">Day</option>
        </select>
      </div>

      <div class="ml-auto flex items-center gap-2">
        <button @click="reload" class="px-3 py-1 bg-blue-600 text-white rounded">새로고침</button>
      </div>
    </div>

    <div v-if="loading" class="text-gray-600">로딩 중...</div>

    <div class="chart-area">
      <canvas ref="canvas" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, computed } from 'vue';
import { getCampaignStats } from '../api';
import {
  Chart,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  LineController,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';

Chart.register(CategoryScale, LinearScale, PointElement, LineElement, LineController, Title, Tooltip, Legend);

const props = defineProps<{ campaignId: string }>();

const canvas = ref<HTMLCanvasElement | null>(null);
let chart: Chart | null = null;

const loading = ref(false);

// Controls
const preset = ref<'24h' | '7d' | '30d' | 'custom'>('24h');
const granularity = ref<'hour' | 'day'>('hour');

// timestamps in seconds
const fromTs = ref<number>(Math.floor(Date.now() / 1000) - 24 * 3600);
const toTs = ref<number>(Math.floor(Date.now() / 1000));

// datetime-local bindings (YYYY-MM-DDTHH:mm)
const fromLocal = ref<string>(formatDateTimeLocal(fromTs.value));
const toLocal = ref<string>(formatDateTimeLocal(toTs.value));

function formatDateTimeLocal(tsSeconds: number) {
  const d = new Date(tsSeconds * 1000);
  const pad = (n: number) => String(n).padStart(2, '0');
  const yyyy = d.getFullYear();
  const mm = pad(d.getMonth() + 1);
  const dd = pad(d.getDate());
  const hh = pad(d.getHours());
  const min = pad(d.getMinutes());
  return `${yyyy}-${mm}-${dd}T${hh}:${min}`;
}

function parseDateTimeLocal(v: string) {
  // interpret as local datetime
  const d = new Date(v);
  return Math.floor(d.getTime() / 1000);
}

function applyPreset() {
  const now = Math.floor(Date.now() / 1000);
  if (preset.value === '24h') {
    fromTs.value = now - 24 * 3600;
    toTs.value = now;
  } else if (preset.value === '7d') {
    fromTs.value = now - 7 * 24 * 3600;
    toTs.value = now;
  } else if (preset.value === '30d') {
    fromTs.value = now - 30 * 24 * 3600;
    toTs.value = now;
  } else if (preset.value === 'custom') {
    // keep current from/toLocal values
    fromTs.value = parseDateTimeLocal(fromLocal.value);
    toTs.value = parseDateTimeLocal(toLocal.value);
  }
  // update local inputs
  fromLocal.value = formatDateTimeLocal(fromTs.value);
  toLocal.value = formatDateTimeLocal(toTs.value);
  reload();
}

function onCustomDateChange() {
  // ensure custom preset selected
  if (preset.value !== 'custom') preset.value = 'custom';
  fromTs.value = parseDateTimeLocal(fromLocal.value);
  toTs.value = parseDateTimeLocal(toLocal.value);
  // don't immediately reload on each change? We'll reload once inputs are valid.
  reload();
}

async function loadAndRender() {
  if (!props.campaignId) return;
  loading.value = true;
  const params: any = {
    granularity: granularity.value,
    from: fromTs.value,
    to: toTs.value,
  };

  const resp: any = await getCampaignStats(props.campaignId, params).catch((e: unknown) => {
    console.error('Failed to load stats', e);
    loading.value = false;
    return null;
  });
  loading.value = false;
  if (!resp) return;

  const labels: string[] = (resp.labels || []).map((ts: number) =>
    new Date(ts * 1000).toLocaleString()
  );
  // For single-campaign endpoint, series should contain one entry
  const series = resp.series && resp.series.length > 0 ? resp.series[0] : null;
  const opens = series ? series.opens : [];
  const clicks = series ? series.clicks : [];

  const data = {
    labels,
    datasets: [
      {
        label: 'Opens',
        data: opens,
        borderColor: '#2563EB',
        backgroundColor: 'rgba(37,99,235,0.1)',
        tension: 0.2,
      },
      {
        label: 'Clicks',
        data: clicks,
        borderColor: '#10B981',
        backgroundColor: 'rgba(16,185,129,0.1)',
        tension: 0.2,
      },
    ],
  };

  if (!canvas.value) {
    console.log('no canvas');
    return;
  }
  const ctx = canvas.value.getContext('2d');
  if (!ctx) {
    console.log('no ctx');
    return;
  }

  if (chart) {
    chart.data = data as any;
    chart.update();
    return;
  }

  chart = new Chart(ctx, {
    type: 'line',
    data: data as any,
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'top',
        },
        title: {
          display: true,
          text: 'Campaign Opens / Clicks',
        },
      },
      scales: {
        x: {
          title: {
            display: true,
            text: 'Time',
          },
        },
        y: {
          title: {
            display: true,
            text: 'Count',
          },
          beginAtZero: true,
        },
      },
    },
  });
}

function reload() {
  // basic validation
  if (fromTs.value >= toTs.value) {
    console.warn('from must be before to');
    return;
  }
  loadAndRender();
}

onMounted(() => {
  applyPreset();
});

onBeforeUnmount(() => {
  if (chart) {
    chart.destroy();
    chart = null;
  }
});

// reload when campaignId or granularity or from/to change
watch(() => props.campaignId, () => {
  applyPreset();
});
watch(granularity, () => {
  reload();
});
watch([fromTs, toTs], () => {
  // sync local values
  fromLocal.value = formatDateTimeLocal(fromTs.value);
  toLocal.value = formatDateTimeLocal(toTs.value);
});
</script>

<style scoped>
.chart-area {
  height: 400px;
  position: relative;
}
canvas {
  width: 100%;
  height: 100% !important;
  display: block;
}
</style>
