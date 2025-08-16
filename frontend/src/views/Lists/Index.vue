<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">구독자 리스트</h1>
        <p class="text-gray-600 mt-1">이메일 구독자 리스트를 관리하고 구성하세요</p>
      </div>
      <button 
        @click="showCreateModal = true" 
        class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-indigo-500 to-purple-600 text-white font-medium rounded-xl hover:from-indigo-600 hover:to-purple-700 transition-all duration-200 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
        </svg>
        새 리스트 만들기
      </button>
    </div>

    <!-- Search -->
    <div class="bg-white rounded-2xl shadow-sm border border-gray-200 p-6">
      <div class="relative">
        <svg class="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
        </svg>
        <input 
          v-model="searchQuery" 
          @input="fetchLists" 
          type="text" 
          placeholder="리스트 검색..." 
          class="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all duration-200"
        >
      </div>
    </div>

    <!-- Lists Grid -->
    <div v-if="lists.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="list in lists" 
        :key="list.id"
        class="bg-white rounded-2xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-200 overflow-hidden group">
        
        <!-- List Header -->
        <div class="p-6 border-b border-gray-100">
          <div class="flex items-start justify-between mb-3">
            <div class="flex-1">
              <h3 class="text-lg font-semibold text-gray-900 mb-2">{{ list.name }}</h3>
              <p class="text-sm text-gray-600 overflow-hidden" style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;">{{ list.description || '설명이 없습니다' }}</p>
              <div class="mt-3">
                <div v-if="list.tags && list.tags.length > 0" class="flex flex-wrap gap-2 mt-2">
                  <span v-for="tag in list.tags" :key="tag" class="text-xs px-2 py-1 bg-indigo-100 text-indigo-700 rounded-full">{{ tag }}</span>
                </div>
              </div>
            </div>
          </div>
          <p class="text-xs text-gray-500">
            {{ list.created_at ? new Date(list.created_at * 1000).toLocaleDateString('ko-KR') : '' }}
          </p>
        </div>

        <!-- List Stats -->
        <div class="p-6 bg-gradient-to-br from-gray-50 to-indigo-50">
          <div class="flex items-center justify-between mb-4">
            <div class="text-center">
              <div class="text-2xl font-bold text-indigo-600">{{ list.subscriber_count || 0 }}</div>
              <div class="text-sm text-gray-600">구독자</div>
            </div>
            <div class="w-16 h-16 bg-gradient-to-br from-indigo-100 to-purple-100 rounded-full flex items-center justify-center">
              <svg class="w-8 h-8 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
              </svg>
            </div>
          </div>
          
          <!-- Actions -->
          <div class="flex justify-between items-center">
              <div class="flex space-x-2">
              <button 
                @click="manageSubscribers(list)"
                class="inline-flex items-center px-3 py-2 text-sm font-medium text-sky-600 bg-sky-50 rounded-lg hover:bg-sky-100 transition-colors duration-200">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7h18M3 12h18M3 17h18"></path>
                </svg>
                구독자 관리
              </button>
              <button 
                @click="editList(list)"
                class="inline-flex items-center px-3 py-2 text-sm font-medium text-indigo-600 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors duration-200">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                </svg>
                편집
              </button>
              <button 
                @click="confirmDelete(list)"
                class="inline-flex items-center px-3 py-2 text-sm font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 transition-colors duration-200">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                </svg>
                삭제
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
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-gray-900 mb-2">아직 리스트가 없습니다</h3>
      <p class="text-gray-600 mb-6">첫 번째 구독자 리스트를 만들어 시작해보세요</p>
      <button 
        @click="showCreateModal = true"
        class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-indigo-500 to-purple-600 text-white font-medium rounded-xl hover:from-indigo-600 hover:to-purple-700 transition-all duration-200">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
        </svg>
        새 리스트 만들기
      </button>
    </div>

    <!-- Pagination -->
    <Pagination
      v-if="pagination && lists.length > 0"
      :current-page="pagination.page || 1"
      :total="pagination.total || 0"
      :limit="pagination.limit || 20"
      @page-change="handlePageChange"
    />

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || editingList" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md overflow-hidden">
        <!-- Modal Header -->
        <div class="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-indigo-50 to-purple-50">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-bold text-gray-900">
                {{ editingList ? '리스트 편집' : '새 리스트 만들기' }}
              </h2>
              <p class="text-gray-600 text-sm mt-1">구독자들을 그룹화할 리스트를 만드세요</p>
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
          <form @submit.prevent="saveList" class="space-y-4">
            <div>
              <label for="name" class="block text-sm font-semibold text-gray-900 mb-2">
                리스트 이름 <span class="text-red-500">*</span>
              </label>
              <input 
                v-model="listForm.name" 
                type="text" 
                id="name" 
                class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all duration-200"
                placeholder="예: VIP 고객, 뉴스레터 구독자 등"
                required>
            </div>
            <div>
              <label for="description" class="block text-sm font-semibold text-gray-900 mb-2">
                설명
              </label>
              <textarea 
                v-model="listForm.description" 
                id="description" 
                rows="3" 
                class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all duration-200"
                placeholder="이 리스트에 대한 간단한 설명을 입력하세요..."></textarea>
            </div>

            <div>
              <label for="tags" class="block text-sm font-semibold text-gray-900 mb-2">
                태그 (쉼표로 구분)
              </label>
              <input
                v-model="listForm.tags"
                type="text"
                id="tags"
                class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all duration-200"
                placeholder="예: vip, korea, newsletter"
              />
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
            @click="saveList"
            class="px-4 py-2 bg-gradient-to-r from-indigo-500 to-purple-600 text-white font-medium rounded-lg hover:from-indigo-600 hover:to-purple-700 transition-all duration-200">
            {{ editingList ? '업데이트' : '생성' }}
          </button>
        </div>
      </div>
    </div>
    <SubscriberPickerModal
    :modelValue="showSubscriberModal"
    :initialSelected="modalInitialSelected"
    @update:modelValue="val => showSubscriberModal = val"
    @confirmed="onSubscriberPickerConfirmed"
  />
</div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { getLists, createList, updateList, deleteList, getSubscribersOfList, patchListSubscribers } from '../../api';
import type { List, PaginationResponse } from '../../types';
import Pagination from '../../components/Pagination.vue';
import SubscriberPickerModal from '../../components/SubscriberPickerModal.vue';

const lists = ref<List[]>([]);
const pagination = ref<PaginationResponse | null>(null);
const searchQuery = ref('');
const currentPage = ref(1);
const limit = ref(20);
const showCreateModal = ref(false);
const editingList = ref<List | null>(null);

const listForm = reactive({
  name: '',
  description: '',
  tags: '',
});

const fetchLists = async () => {
  try {
    const response = await getLists({ 
      page: currentPage.value,
      limit: limit.value,
      search: searchQuery.value || undefined 
    });
    
    if (response && typeof response === 'object' && 'data' in response) {
      const paginatedResponse = response as any;
      lists.value = paginatedResponse.data || [];
      pagination.value = paginatedResponse.pagination || null;
    }
  } catch (error) {
    console.error('Failed to fetch lists:', error);
  }
};

const saveList = async () => {
  try {
    // normalize tags: comma-separated string -> string[]
    const tagsArray = (listForm.tags || '')
      .split(',')
      .map(t => t.trim())
      .filter(Boolean);

    const payload: any = {
      name: listForm.name,
      description: listForm.description,
      tags: tagsArray.length > 0 ? tagsArray : undefined,
    };

    if (editingList.value && editingList.value.id) {
      await updateList(editingList.value.id, payload);
    } else {
      await createList(payload);
    }
    fetchLists();
    closeModal();
  } catch (error) {
    console.error('Failed to save list:', error);
  }
};

const editList = (list: List) => {
  editingList.value = list;
  listForm.name = list.name || '';
  listForm.description = list.description || '';
  listForm.tags = (list.tags && Array.isArray(list.tags)) ? list.tags.join(', ') : '';
  showCreateModal.value = true;
};

const confirmDelete = async (list: List) => {
  if (list.id && window.confirm(`"${list.name}" 리스트를 정말 삭제하시겠습니까?`)) {
    try {
      await deleteList(list.id);
      fetchLists();
    } catch (error) {
      console.error('Failed to delete list:', error);
    }
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchLists();
};

const closeModal = () => {
  showCreateModal.value = false;
  editingList.value = null;
  listForm.name = '';
  listForm.description = '';
  listForm.tags = '';
};

//
// Subscriber management modal for a list
//
const showSubscriberModal = ref(false);
const modalListId = ref<string | null>(null);
const modalInitialSelected = ref<string[] | null>(null);

// manageSubscribers opens the subscriber picker for a list, preloading current subscribers.
const manageSubscribers = async (list: List) => {
  modalListId.value = list.id;
  try {
    const res = await getSubscribersOfList(list.id, { page: 1, limit: 1000 });
    if (res && typeof res === 'object' && 'data' in res) {
      const data = (res as any).data || [];
      modalInitialSelected.value = data.map((s: any) => s.id);
    } else {
      modalInitialSelected.value = [];
    }
  } catch (err) {
    console.error('Failed to load list subscribers', err);
    modalInitialSelected.value = [];
  }
  showSubscriberModal.value = true;
};

// onSubscriberPickerConfirmed applies changes (add/remove) using PATCH endpoint.
const onSubscriberPickerConfirmed = async (selected: any[]) => {
  if (!modalListId.value) return;
  const selectedIds = selected.map(s => s.id);
  const currentIds = modalInitialSelected.value || [];
  const toAdd = selectedIds.filter((id: string) => !currentIds.includes(id));
  const toRemove = currentIds.filter((id: string) => !selectedIds.includes(id));

  try {
    if (toRemove.length > 0) {
      await patchListSubscribers(modalListId.value, { add: [], remove: toRemove });
    }
    if (toAdd.length > 0) {
      await patchListSubscribers(modalListId.value, { add: toAdd, remove: [] });
    }
    // refresh list counts
    fetchLists();
  } catch (err) {
    console.error('Failed to update list subscribers', err);
  } finally {
    showSubscriberModal.value = false;
    modalListId.value = null;
    modalInitialSelected.value = null;
  }
};

onMounted(fetchLists);
</script>
