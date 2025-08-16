<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">구독자 관리</h1>
        <p class="text-gray-600 mt-1">이메일 구독자를 관리하고 세그먼트를 구성하세요</p>
      </div>
      <button 
        @click="showCreateModal = true" 
        class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-blue-500 to-cyan-600 text-white font-medium rounded-xl hover:from-blue-600 hover:to-cyan-700 transition-all duration-200 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
        </svg>
        새 구독자 추가
      </button>
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
              @input="fetchSubscribers" 
              type="text" 
              placeholder="구독자 검색..." 
              class="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
            >
          </div>
        </div>
        <div class="flex gap-2">
          <select 
            v-model="selectedListId" 
            @change="fetchSubscribers"
            class="px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200">
            <option value="">모든 리스트</option>
            <option v-for="list in lists" :key="list.id" :value="list.id">{{ list.name }}</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Subscribers Table -->
    <div v-if="subscribers.length > 0" class="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">구독자</th>
              <th class="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">상태</th>
              <th class="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">리스트</th>
              <th class="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">가입일</th>
              <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">작업</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="subscriber in subscribers" :key="subscriber.id" class="hover:bg-gray-50 transition-colors duration-200">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div class="w-10 h-10 bg-gradient-to-br from-blue-100 to-cyan-100 rounded-full flex items-center justify-center">
                    <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                    </svg>
                  </div>
                  <div class="ml-4">
                    <div class="text-sm font-semibold text-gray-900">{{ subscriber.name || '이름 없음' }}</div>
                    <div class="text-sm text-gray-600">{{ subscriber.email }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getStatusBadgeClass(subscriber.status)" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
                  {{ getStatusText(subscriber.status) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-900">{{ getListNames(subscriber.lists) }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                {{ subscriber.created_at ? new Date(subscriber.created_at * 1000).toLocaleDateString('ko-KR') : '' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <div class="flex justify-end space-x-2">
                  <button 
                    @click="editSubscriber(subscriber)"
                    class="inline-flex items-center px-3 py-1.5 text-xs font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors duration-200">
                    <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                    </svg>
                    편집
                  </button>
                  <button 
                    @click="confirmDelete(subscriber)"
                    class="inline-flex items-center px-3 py-1.5 text-xs font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 transition-colors duration-200">
                    <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                    </svg>
                    삭제
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-16">
      <div class="w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-6">
        <svg class="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-gray-900 mb-2">아직 구독자가 없습니다</h3>
      <p class="text-gray-600 mb-6">첫 번째 구독자를 추가해 시작해보세요</p>
      <button 
        @click="showCreateModal = true"
        class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-blue-500 to-cyan-600 text-white font-medium rounded-xl hover:from-blue-600 hover:to-cyan-700 transition-all duration-200">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
        </svg>
        새 구독자 추가
      </button>
    </div>

    <!-- Pagination -->
    <Pagination
      v-if="pagination && subscribers.length > 0"
      :current-page="pagination.page || 1"
      :total="pagination.total || 0"
      :limit="pagination.limit || 20"
      @page-change="handlePageChange"
    />

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || editingSubscriber" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md overflow-hidden">
        <!-- Modal Header -->
        <div class="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-blue-50 to-cyan-50">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-bold text-gray-900">
                {{ editingSubscriber ? '구독자 편집' : '새 구독자 추가' }}
              </h2>
              <p class="text-gray-600 text-sm mt-1">구독자 정보를 입력하세요</p>
            </div>
            <button 
              @click="closeModal"
              class="p-2 hover:bg-gray-100 rounded-full transition-colors duration-200">
              <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>
        </div>

        <!-- Modal Body -->
        <div class="p-6">
          <form @submit.prevent="saveSubscriber" class="space-y-4">
            <div>
              <label for="email" class="block text-sm font-semibold text-gray-900 mb-2">
                이메일 주소 <span class="text-red-500">*</span>
              </label>
              <input 
                v-model="subscriberForm.email" 
                type="email" 
                id="email" 
                class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                placeholder="subscriber@example.com"
                required>
            </div>
            <div>
              <label for="name" class="block text-sm font-semibold text-gray-900 mb-2">
                이름
              </label>
              <input 
                v-model="subscriberForm.name" 
                type="text" 
                id="name" 
                class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                placeholder="구독자 이름">
            </div>
            <div v-if="!editingSubscriber">
              <label for="list_id" class="block text-sm font-semibold text-gray-900 mb-2">
                리스트 <span class="text-red-500">*</span>
              </label>
              <select 
                v-model="subscriberForm.list_id" 
                id="list_id" 
                class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                required>
                <option value="">리스트를 선택하세요</option>
                <option v-for="list in lists" :key="list.id" :value="list.id">{{ list.name }}</option>
              </select>
            </div>
            <div v-if="editingSubscriber">
              <label for="status" class="block text-sm font-semibold text-gray-900 mb-2">
                상태 <span class="text-red-500">*</span>
              </label>
              <select 
                v-model="subscriberForm.status" 
                id="status" 
                class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                required>
                <option value="enabled">활성</option>
                <option value="disabled">비활성</option>
              </select>
            </div>
          </form>
        </div>

        <!-- Modal Footer -->
        <div class="px-6 py-4 border-t border-gray-200 bg-gray-50 flex justify-end space-x-3">
          <button 
            type="button" 
            @click="closeModal" 
            class="px-4 py-2 border border-gray-300 text-gray-700 font-medium rounded-lg hover:bg-gray-100 transition-all duration-200">
            취소
          </button>
          <button 
            @click="saveSubscriber"
            class="px-4 py-2 bg-gradient-to-r from-blue-500 to-cyan-600 text-white font-medium rounded-lg hover:from-blue-600 hover:to-cyan-700 transition-all duration-200">
            {{ editingSubscriber ? '업데이트' : '추가' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { getSubscribers, updateSubscriber, deleteSubscriber, getLists, patchListSubscribers } from '../../api';
import type { Subscriber, List, PaginationResponse } from '../../types';
import Pagination from '../../components/Pagination.vue';
import ListPickerModal from '../../components/ListPickerModal.vue';

const subscribers = ref<Subscriber[]>([]);
const lists = ref<List[]>([]);
const pagination = ref<PaginationResponse | null>(null);
const searchQuery = ref('');
const selectedListId = ref('');
const currentPage = ref(1);
const limit = ref(20);
const showCreateModal = ref(false);
const editingSubscriber = ref<Subscriber | null>(null);

const subscriberForm = reactive({
  email: '',
  name: '',
  status: 'enabled' as string,
  list_id: '',
});

const fetchSubscribers = async () => {
  try {
    const response = await getSubscribers({ 
      page: currentPage.value,
      limit: limit.value,
      search: searchQuery.value || undefined
    });
    
    if (response && typeof response === 'object' && 'data' in response) {
      const paginatedResponse = response as any;
      subscribers.value = paginatedResponse.data || [];
      pagination.value = paginatedResponse.pagination || null;
    }
  } catch (error) {
    console.error('Failed to fetch subscribers:', error);
  }
};

const fetchLists = async () => {
  try {
    const response = await getLists({ page: 1, limit: 100 });
    if (response && typeof response === 'object' && 'data' in response) {
      const paginatedResponse = response as any;
      lists.value = paginatedResponse.data || [];
    }
  } catch (error) {
    console.error('Failed to fetch lists:', error);
  }
};

const saveSubscriber = async () => {
  try {
    if (editingSubscriber.value && editingSubscriber.value.id) {
      const { list_id, ...updateData } = subscriberForm;
      await updateSubscriber(editingSubscriber.value.id, updateData);
    } else {
      // Create functionality not available in API yet
      console.log('Create subscriber functionality not implemented');
    }
    fetchSubscribers();
    closeModal();
  } catch (error) {
    console.error('Failed to save subscriber:', error);
  }
};

const editSubscriber = (subscriber: Subscriber) => {
  editingSubscriber.value = subscriber;
  subscriberForm.email = subscriber.email || '';
  subscriberForm.name = subscriber.name || '';
  subscriberForm.status = subscriber.status || 'enabled';
  showCreateModal.value = true;
};

const confirmDelete = async (subscriber: Subscriber) => {
  if (subscriber.id && window.confirm(`"${subscriber.email}" 구독자를 정말 삭제하시겠습니까?`)) {
    try {
      await deleteSubscriber(subscriber.id);
      fetchSubscribers();
    } catch (error) {
      console.error('Failed to delete subscriber:', error);
    }
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchSubscribers();
};

const closeModal = () => {
  showCreateModal.value = false;
  editingSubscriber.value = null;
  subscriberForm.email = '';
  subscriberForm.name = '';
  subscriberForm.status = 'enabled';
  subscriberForm.list_id = '';
};

//
// List management modal for a subscriber
//
const showListModal = ref(false);
const modalSubscriberId = ref<string | null>(null);
const modalInitialSelectedLists = ref<string[] | null>(null);

const manageLists = (subscriber: Subscriber) => {
  modalSubscriberId.value = subscriber.id || null;
  modalInitialSelectedLists.value = (subscriber.lists || []).map((l: any) => l.list_id);
  showListModal.value = true;
};

const onListPickerConfirmed = async (selected: any[]) => {
  if (!modalSubscriberId.value) return;
  const subscriberId = modalSubscriberId.value;
  const selectedIds = selected.map((s: any) => s.id);
  const currentIds = modalInitialSelectedLists.value || [];

  const toAdd = selectedIds.filter((id: string) => !currentIds.includes(id));
  const toRemove = currentIds.filter((id: string) => !selectedIds.includes(id));

  try {
    // For each list to add, call patchListSubscribers to add this subscriber
    for (const listID of toAdd) {
      await patchListSubscribers(listID, { add: [subscriberId], remove: [] });
    }
    // For each list to remove, call patchListSubscribers to remove this subscriber
    for (const listID of toRemove) {
      await patchListSubscribers(listID, { add: [], remove: [subscriberId] });
    }
    await fetchSubscribers();
  } catch (err) {
    console.error('Failed to update subscriber lists', err);
  } finally {
    showListModal.value = false;
    modalSubscriberId.value = null;
    modalInitialSelectedLists.value = null;
  }
};

const getStatusBadgeClass = (status: string | undefined) => {
  switch (status) {
    case 'enabled':
      return 'bg-green-100 text-green-800';
    case 'disabled':
      return 'bg-gray-100 text-gray-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
};

const getStatusText = (status: string | undefined) => {
  switch (status) {
    case 'enabled':
      return '활성';
    case 'disabled':
      return '비활성';
    default:
      return status || '알 수 없음';
  }
};

const getListName = (listId: string | undefined) => {
  if (!listId) return '리스트 없음';
  const list = lists.value.find((l: List) => l.id === listId);
  return list?.name || '알 수 없음';
};

const getListNames = (subscriberLists: any[] | undefined) => {
  if (!subscriberLists || subscriberLists.length === 0) return '리스트 없음';
  
  const listNames = subscriberLists
    .map((sl: any) => {
      const list = lists.value.find((li: List) => li.id === sl.list_id);
      return list?.name || '알 수 없음';
    })
    .filter((name: string) => name !== '알 수 없음');
  
  return listNames.length > 0 ? listNames.join(', ') : '리스트 없음';
};

onMounted(() => {
  fetchSubscribers();
  fetchLists();
});
</script>
