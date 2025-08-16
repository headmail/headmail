<template>
  <div class="bg-white rounded-2xl shadow-sm border border-gray-200 p-6">
    <canvas ref="canvas" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue';
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

async function loadAndRender() {
  if (!props.campaignId) return;
  const resp: any = await getCampaignStats(props.campaignId, { granularity: 'hour' }).catch((e: unknown) => {
    console.error('Failed to load stats', e);
    return null;
  });
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

  if (!canvas.value) return;
  const ctx = canvas.value.getContext('2d');
  if (!ctx) return;

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

onMounted(() => {
  loadAndRender();
});

onBeforeUnmount(() => {
  if (chart) {
    chart.destroy();
    chart = null;
  }
});

// reload when campaignId changes
watch(() => props.campaignId, () => {
  loadAndRender();
});
</script>

<style scoped>
canvas {
  width: 100%;
  height: 320px;
}
</style>
