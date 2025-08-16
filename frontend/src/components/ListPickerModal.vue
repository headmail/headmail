<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center">
    <div class="absolute inset-0 bg-black/50" @click="close"></div>
    <div class="relative bg-white w-full max-w-2xl mx-4 rounded-2xl shadow-xl border border-gray-200 p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900">리스트 선택</h2>
        <button @click="close" class="p-2 text-gray-400 hover:text-gray-600">
          <span class="sr-only">닫기</span>
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <div class="space-y-3">
        <div class="flex gap-2">
          <input v-model="search" @input="fetch(1)" type="text" placeholder="리스트 검색..." class="flex-1 px-3 py-2 border rounded" />
          <select v-model.number="limit" @change="fetch(1)" class="px-3 py-2 border rounded">
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="50">50</option>
          </select>
        </div>

        <div class="max-h-72 overflow-auto border rounded p-2">
          <div v-if="lists.length === 0" class="text-sm text-gray-500">리스트가 없습니다.</div>
          <label v-for="list in lists" :key="list.id" class="flex items-center justify-between gap-2 p-2 hover:bg-gray-50 rounded">
            <div class="flex items-center gap-2">
              <input type="checkbox" :value="list.id" v-model="selectedLocal" />
              <div>
                <div class="text-sm font-medium text-gray-900">{{ list.name }}</div>
                <div class="text-xs text-gray-500">{{ list.description || '' }}</div>
              </div>
            </div>
            <div class="text-xs text-gray-400">{{ list.subscriber_count || 0 }}</div>
          </label>
        </div>

        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-600">총: {{ pagination.total || 0 }}</div>
          <div class="flex items-center gap-2">
            <button @click="prevPage" :disabled="pagination.page <= 1" class="px-3 py-1 border rounded disabled:opacity-50">이전</button>
            <div class="px-3 py-1 border rounded">{{ pagination.page || 1 }}</div>
            <button @click="nextPage" :disabled="pagination.page >= Math.ceil((pagination.total || 0) / (pagination.limit || 10))" class="px-3 py-1 border rounded disabled:opacity-50">다음</button>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button @click="close" type="button" class="px-4 py-2 rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50">취소</button>
          <button @click="confirm" type="button" class="px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700">선택 완료</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { getLists } from '../api';
import type { List } from '../types';

const props = defineProps<{
  modelValue: boolean;
  initialSelected?: string[] | null;
}>();

const emit = defineEmits(['update:modelValue', 'confirmed']);

const visible = ref<boolean>(props.modelValue || false);
const lists = ref<List[]>([]);
const selectedLocal = ref<string[]>(props.initialSelected ? [...props.initialSelected] : []);
const search = ref<string>('');
const page = ref<number>(1);
const limit = ref<number>(10);
const pagination = ref<{ page: number; limit: number; total: number }>({ page: 1, limit: 10, total: 0 });

watch(() => props.modelValue, (v) => {
  visible.value = v;
  if (v) {
    fetch(1);
  }
});

watch(() => props.initialSelected, (v) => {
  selectedLocal.value = v ? [...v] : [];
});

const close = () => {
  emit('update:modelValue', false);
};

const fetch = async (p = 1) => {
  page.value = p;
  try {
    const resp = await getLists({ page: page.value, limit: limit.value, search: search.value || undefined });
    if (resp && typeof resp === 'object' && 'data' in resp) {
      const r = resp as any;
      lists.value = r.data || [];
      pagination.value = r.pagination || { page: 1, limit: limit.value, total: 0 };
    } else {
      lists.value = [];
      pagination.value = { page: 1, limit: limit.value, total: 0 };
    }
  } catch (err) {
    console.error('리스트 로드 실패', err);
    lists.value = [];
    pagination.value = { page: 1, limit: limit.value, total: 0 };
  }
};

const prevPage = () => {
  if (page.value > 1) fetch(page.value - 1);
};
const nextPage = () => {
  const last = Math.ceil((pagination.value.total || 0) / (pagination.value.limit || limit.value));
  if (page.value < last) fetch(page.value + 1);
};

const confirm = () => {
  const selectedObjs = selectedLocal.value
    .map((id) => lists.value.find((l: List) => l.id === id))
    .filter(Boolean) as List[];
  emit('confirmed', selectedObjs);
  close();
};
</script>
