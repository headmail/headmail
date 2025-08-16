<template>
  <div class="space-y-6">
    <div v-if="deliveries.length === 0" class="text-center py-16 text-gray-600">
      전송 항목이 없습니다.
    </div>

    <div v-else class="bg-white rounded-2xl shadow-sm border border-gray-200 p-4">
      <table class="w-full text-sm">
        <thead>
        <tr class="text-left text-gray-500">
          <th class="px-3 py-2">Email</th>
          <th class="px-3 py-2">Name</th>
          <th class="px-3 py-2">Status</th>
          <th class="px-3 py-2">Sent At</th>
          <th class="px-3 py-2">Opens</th>
          <th class="px-3 py-2">Clicks</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="d in deliveries" :key="d.id" class="border-t">
          <td class="px-3 py-2">{{ d.email }}</td>
          <td class="px-3 py-2">{{ d.name || '-' }}</td>
          <td class="px-3 py-2">{{ d.status }}</td>
          <td class="px-3 py-2">{{ d.sent_at ? new Date(d.sent_at * 1000).toLocaleString() : '-' }}</td>
          <td class="px-3 py-2">{{ d.open_count ?? 0 }}</td>
          <td class="px-3 py-2">{{ d.click_count ?? 0 }}</td>
        </tr>
        </tbody>
      </table>

      <div class="mt-4 flex justify-between items-center">
        <div class="text-sm text-gray-600">총: {{ total }}</div>
        <div class="flex items-center space-x-2">
          <button class="px-3 py-1 border rounded" :disabled="page <= 1" @click="prevPage">이전</button>
          <div class="text-sm">페이지 {{ page }}</div>
          <button class="px-3 py-1 border rounded" :disabled="(page * limit) >= total" @click="nextPage">다음</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { getCampaignDeliveries } from '../api';

const props = defineProps<{ campaignId?: string }>();
const route = useRoute();
const campaignId = String(props.campaignId ?? route.params.id ?? '');

const deliveries = ref<any[]>([]);
const total = ref(0);
const page = ref(1);
const limit = ref(20);
const loading = ref(false);

const fetchDeliveries = async () => {
  loading.value = true;
  try {
    const resp = await getCampaignDeliveries(campaignId, { page: page.value, limit: limit.value });
    if (resp && resp.data) {
      deliveries.value = resp.data;
      total.value = resp.pagination?.total ?? 0;
    } else if (Array.isArray(resp)) {
      // fallback if API returns array
      deliveries.value = resp;
      total.value = resp.length;
    }
  } catch (err) {
    console.error('Failed to fetch deliveries', err);
  } finally {
    loading.value = false;
  }
};

const prevPage = () => {
  if (page.value > 1) {
    page.value--;
    fetchDeliveries();
  }
};
const nextPage = () => {
  if (page.value * limit.value < total.value) {
    page.value++;
    fetchDeliveries();
  }
};

onMounted(fetchDeliveries);
</script>

<style scoped>
table th, table td {
  vertical-align: middle;
}
</style>
