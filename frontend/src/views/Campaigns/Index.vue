<template>
  <div>
    <h1 class="text-2xl font-bold mb-4">Campaigns</h1>
    <!-- Search and filter controls -->
    <div class="mb-4">
      <input 
        v-model="searchQuery" 
        @input="fetchCampaigns" 
        type="text" 
        placeholder="Search campaigns..." 
        class="p-2 border rounded w-full"
      >
    </div>
    <table class="min-w-full bg-white">
      <thead>
        <tr>
          <th class="py-2 px-4 border-b">Name</th>
          <th class="py-2 px-4 border-b">Subject</th>
          <th class="py-2 px-4 border-b">Status</th>
          <th class="py-2 px-4 border-b">Created At</th>
          <th class="py-2 px-4 border-b">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="campaign in campaigns" :key="campaign.id">
          <td class="py-2 px-4 border-b">{{ campaign.name }}</td>
          <td class="py-2 px-4 border-b">{{ campaign.subject }}</td>
          <td class="py-2 px-4 border-b">{{ campaign.status }}</td>
          <td class="py-2 px-4 border-b">{{ campaign.created_at ? new Date(campaign.created_at * 1000).toLocaleString() : '' }}</td>
          <td class="py-2 px-4 border-b">
            <router-link :to="{ name: 'CampaignDetail', params: { id: campaign.id } }" class="text-blue-500 hover:underline">View</router-link>
          </td>
        </tr>
      </tbody>
    </table>
    
    <!-- Pagination -->
    <Pagination
      v-if="pagination"
      :current-page="pagination.page || 1"
      :total="pagination.total || 0"
      :limit="pagination.limit || 20"
      @page-change="handlePageChange"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getCampaigns } from '../../api';
import type { Campaign, PaginationResponse } from '../../types';
import Pagination from '../../components/Pagination.vue';

const campaigns = ref<Campaign[]>([]);
const pagination = ref<PaginationResponse | null>(null);
const searchQuery = ref('');
const currentPage = ref(1);
const limit = ref(20);

const fetchCampaigns = async () => {
  try {
    const response = await getCampaigns({
      page: currentPage.value,
      limit: limit.value,
      search: searchQuery.value || undefined
    });
    
    if (response && typeof response === 'object' && 'data' in response) {
      const paginatedResponse = response as any;
      campaigns.value = paginatedResponse.data || [];
      pagination.value = paginatedResponse.pagination || null;
    }
  } catch (error) {
    console.error('Failed to fetch campaigns:', error);
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchCampaigns();
};

onMounted(fetchCampaigns);
</script>
