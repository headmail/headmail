<template>
  <div>
    <h1 class="text-2xl font-bold mb-4">Create New Campaign</h1>
    <div class="mb-4">
      <label for="campaign-name" class="block text-sm font-medium text-gray-700">Name</label>
      <input type="text" id="campaign-name" v-model="campaign.name" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
    </div>
    <div class="mb-4">
      <label for="campaign-subject" class="block text-sm font-medium text-gray-700">Subject</label>
      <input type="text" id="campaign-subject" v-model="campaign.subject" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
    </div>
    <button @click="handleCreate" class="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-700">Create Campaign</button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { createCampaign } from '../../api';
import type { Campaign } from '../../types';

const router = useRouter();
const campaign = ref({
  name: '',
  subject: '',
});

const handleCreate = async () => {
  try {
    const newCampaign = await createCampaign({
      name: campaign.value.name,
      subject: campaign.value.subject,
    });
    console.log('newCampaign: ', newCampaign);
    // alert('Campaign created successfully!');
    router.push({ name: 'Campaigns' });
    // if (newCampaign && newCampaign.id) {
    //   router.push({ name: 'CampaignDetail', params: { id: newCampaign.id } });
    // } else {
    //   router.push({ name: 'Campaigns' });
    // }
  } catch (error) {
    console.error('Failed to create campaign:', error);
    alert('Failed to create campaign.');
  }
};
</script>
