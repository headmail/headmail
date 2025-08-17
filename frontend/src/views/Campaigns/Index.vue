<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">이메일 캠페인</h1>
        <p class="text-gray-600 mt-1">이메일 마케팅 캠페인을 관리하고 성과를 확인하세요</p>
      </div>
      <router-link :to="{ name: 'CampaignCreate' }" class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-green-500 to-emerald-600 text-white font-medium rounded-xl hover:from-green-600 hover:to-emerald-700 transition-all duration-200 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
        </svg>
        새 캠페인 만들기
      </router-link>
    </div>

    <!-- Search and Filters -->
    <div class="bg-white rounded-2xl shadow-sm border border-gray-200 p-6">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <div class="relative">
            <svg class="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
            </svg>
            <input 
              v-model="searchQuery" 
              @input="fetchCampaigns" 
              type="text" 
              placeholder="캠페인 검색..." 
              class="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            >
          </div>
        </div>
        <div class="flex gap-2">
          <select class="px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200">
            <option>모든 상태</option>
            <option>초안</option>
            <option>발송 중</option>
            <option>완료</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Campaigns Grid -->
    <div v-if="campaigns.length > 0" class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
      <div 
        v-for="campaign in campaigns" 
        :key="campaign.id"
        class="bg-white rounded-2xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-200 overflow-hidden group">
        <!-- Campaign Header -->
        <div class="p-6 border-b border-gray-100">
          <div class="flex items-start justify-between mb-3">
            <h3 class="text-lg font-semibold text-gray-900 truncate flex-1 mr-3">{{ campaign.name }}</h3>
            <span :class="getStatusBadgeClass(campaign.status)" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
              {{ getStatusText(campaign.status) }}
            </span>
          </div>
          <p class="text-sm text-gray-600 mb-3 overflow-hidden" style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;">{{ campaign.subject }}</p>
          <p class="text-xs text-gray-500">
            {{ campaign.created_at ? new Date(campaign.created_at * 1000).toLocaleDateString('ko-KR') : '' }}
          </p>
        </div>

        <!-- Campaign Stats -->
        <div class="p-6 bg-gray-50">
          <div class="grid grid-cols-4 gap-4 mb-4">
            <div class="text-center">
              <div class="text-lg font-semibold text-gray-900">{{ campaign.recipient_count }}</div>
              <div class="text-xs text-gray-500">수신자</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-semibold text-gray-900">{{ campaign.delivered_count }}</div>
              <div class="text-xs text-gray-500">발송</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-semibold text-green-600">{{ openRate(campaign) }}%</div>
              <div class="text-xs text-gray-500">열람률</div>
            </div>
            <div class="text-center">
              <div class="text-lg font-semibold text-blue-600">{{ clickRate(campaign) }}%</div>
              <div class="text-xs text-gray-500">클릭률</div>
            </div>
          </div>
          
          <!-- Actions -->
          <div class="flex justify-between items-center">
            <router-link 
              :to="{ name: 'CampaignDetail', params: { id: campaign.id } }" 
              class="inline-flex items-center px-4 py-2 text-sm font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors duration-200">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
              </svg>
              상세보기
            </router-link>
            <div class="flex space-x-2">
              <button @click="deleteCampaign(campaign.id)" class="p-2 text-gray-400 hover:text-red-600 transition-colors duration-200">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-16">
      <div class="w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-6">
        <svg class="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-gray-900 mb-2">아직 캠페인이 없습니다</h3>
      <p class="text-gray-600 mb-6">첫 번째 이메일 캠페인을 만들어 시작해보세요</p>
      <router-link :to="{ name: 'CampaignCreate' }" class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-green-500 to-emerald-600 text-white font-medium rounded-xl hover:from-green-600 hover:to-emerald-700 transition-all duration-200">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
        </svg>
        새 캠페인 만들기
      </router-link>
    </div>
    
    <!-- Pagination -->
    <Pagination
      v-if="pagination && campaigns.length > 0"
      :current-page="pagination.page || 1"
      :total="pagination.total || 0"
      :limit="pagination.limit || 20"
      @page-change="handlePageChange"
    />
    <!-- Edit Modal -->
    <CampaignEditModal
      v-model="isEditOpen"
      :campaign="selectedCampaign"
      @saved="onSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getCampaigns } from '../../api';
import CampaignEditModal from '../../components/CampaignEditModal.vue';
import type { Campaign, PaginationResponse } from '../../types';
import Pagination from '../../components/Pagination.vue';

const campaigns = ref<Campaign[]>([]);
const pagination = ref<PaginationResponse | null>(null);
const searchQuery = ref('');
const currentPage = ref(1);
const limit = ref(20);

const isEditOpen = ref(false);
const selectedCampaign = ref<Campaign | null>(null);

const openEdit = (c: Campaign) => {
  selectedCampaign.value = c;
  isEditOpen.value = true;
};

const onSaved = (updated: any) => {
  // update local list if present
  const idx = campaigns.value.findIndex((it: Campaign) => String(it.id) === String(updated?.id));
  if (idx !== -1) {
    campaigns.value[idx] = { ...campaigns.value[idx], ...(updated as Partial<Campaign>) } as Campaign;
  } else {
    // refetch if not found
    fetchCampaigns();
  }
};

const deleteCampaign = async (id: string | number | undefined) => {
  if (!id) return;
  if (!confirm('정말 삭제하시겠습니까?')) return;
  try {
    await (await import('../../api')).deleteCampaign(String(id));
    campaigns.value = campaigns.value.filter((c: Campaign) => String(c.id) !== String(id));
  } catch (err) {
    console.error('삭제 실패:', err);
    alert('삭제에 실패했습니다.');
  }
};

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

const getStatusBadgeClass = (status: string | undefined) => {
  switch (status?.toLowerCase()) {
    case 'draft':
    case '초안':
      return 'bg-gray-100 text-gray-800';
    case 'scheduled':
    case '예약됨':
      return 'bg-blue-100 text-blue-800';
    case 'sending':
    case '발송중':
      return 'bg-blue-100 text-blue-800';
    case 'sent':
    case '완료':
      return 'bg-green-100 text-green-800';
    case 'failed':
    case '실패':
      return 'bg-red-100 text-red-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
};

const getStatusText = (status: string | undefined) => {
  switch (status?.toLowerCase()) {
    case 'draft':
      return '초안';
    case 'scheduled':
      return '예약됨';
    case 'sending':
      return '발송중';
    case 'sent':
      return '완료';
    case 'failed':
      return '실패';
    default:
      return status || '알 수 없음';
  }
};

const openRate = (c: Campaign | null | undefined) => {
  if (!c) return 0;
  const denom = (c.delivered_count ?? c.recipient_count ?? 0);
  if (!denom) return 0;
  return Math.round(((c.open_count ?? 0) / denom) * 100);
};

const clickRate = (c: Campaign | null | undefined) => {
  if (!c) return 0;
  const denom = (c.delivered_count ?? c.recipient_count ?? 0);
  if (!denom) return 0;
  return Math.round(((c.click_count ?? 0) / denom) * 100);
};

onMounted(fetchCampaigns);
</script>
