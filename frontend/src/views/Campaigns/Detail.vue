<template>
  <div>
    <h1 class="text-2xl font-bold mb-4">Campaign Detail</h1>
    <div v-if="campaign">
      <p><strong>ID:</strong> {{ campaign.id }}</p>
      <p><strong>Name:</strong> {{ campaign.name }}</p>
      <p><strong>Subject:</strong> {{ campaign.subject }}</p>
      <p><strong>Status:</strong> {{ campaign.status }}</p>
      <p><strong>Created At:</strong> {{ new Date(campaign.created_at * 1000).toLocaleString() }}</p>
    </div>
    <div v-else>
      <p>Loading campaign details...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { getCampaign } from '../../api'; // This function needs to be created in api/index.ts
import type { Campaign } from '../../types';

const route = useRoute();
const campaign = ref<Campaign | null>(null);

onMounted(async () => {
  const campaignId = route.params.id as string;
  try {
    const response = await getCampaign(campaignId); // This function needs to be created
    campaign.value = response;
  } catch (error) {
    console.error('Failed to fetch campaign details:', error);
  }
});
</script>
