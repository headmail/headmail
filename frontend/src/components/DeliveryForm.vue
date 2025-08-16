<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div class="space-y-4 bg-white rounded-2xl shadow-sm border border-gray-200 p-6">
    <h2 class="text-lg font-semibold">전송</h2>

    <!-- 선택된 리스트 -->
    <div>
      <div class="flex items-center justify-between mb-2">
        <div class="text-sm text-gray-600">선택된 리스트 (여러개)</div>
        <button @click="openListPicker" class="px-3 py-1 text-sm border rounded">리스트 선택</button>
      </div>
      <div v-if="selectedLists.length === 0" class="text-sm text-gray-500">선택된 리스트가 없습니다.</div>
      <ul v-else class="list-disc pl-5 text-sm">
        <li v-for="l in selectedLists" :key="l.id" class="flex items-center gap-2">
          <span>{{ l.name }}</span>
          <button @click="removeList(l.id)" class="text-xs text-red-600 hover:underline">제거</button>
        </li>
      </ul>
    </div>

    <!-- 개인 수신자 -->
    <div>
      <div class="flex items-center justify-between mb-2">
        <div class="text-sm text-gray-600">개별 수신자 추가</div>
        <div class="flex gap-2">
          <button @click="addIndividual" class="px-3 py-1 text-sm border rounded">행 추가</button>
          <button @click="clearIndividuals" class="px-3 py-1 text-sm border rounded">초기화</button>
        </div>
      </div>

      <div class="space-y-3">
        <div v-for="(ind, idx) in individuals" :key="ind._localId" class="p-3 border rounded">
          <div class="flex justify-between items-start gap-2 mb-2">
            <div class="text-sm font-medium">수신자 {{ idx + 1 }}</div>
            <div class="flex items-center gap-2">
              <button @click="removeIndividual(idx)" class="text-sm text-red-600">삭제</button>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-2">
            <input v-model="ind.email" type="email" placeholder="이메일" class="px-3 py-2 border rounded" />
            <input v-model="ind.name" type="text" placeholder="이름 (선택)" class="px-3 py-2 border rounded" />
            <input v-model="ind.headersText" type="text" placeholder='headers JSON (예: {"X-A":"1"})' class="px-3 py-2 border rounded" />
          </div>

          <div class="mt-2">
            <label class="text-xs text-gray-500">data (JSON)</label>
            <textarea v-model="ind.dataText" rows="3" class="w-full px-3 py-2 border rounded" placeholder='{"foo":"bar"}'></textarea>
            <div v-if="ind.dataError" class="text-xs text-red-600 mt-1">{{ ind.dataError }}</div>
          </div>
          <div v-if="ind.headersError" class="text-xs text-red-600 mt-1">{{ ind.headersError }}</div>
        </div>
      </div>
    </div>

    <!-- 스케줄링 -->
    <div>
      <div class="text-sm text-gray-600 mb-2">스케줄 (비워두면 즉시 전송)</div>
      <input v-model="scheduledAtLocal" type="datetime-local" class="px-3 py-2 border rounded" />
    </div>

    <!-- 전송용 데이터/헤더 전역 입력 (옵션) -->
    <div>
      <div class="text-sm text-gray-600 mb-2">전송 공통 데이터 (JSON, 선택)</div>
      <textarea v-model="commonDataText" rows="3" class="w-full px-3 py-2 border rounded" placeholder='{"global":"value"}'></textarea>
      <div v-if="commonDataError" class="text-xs text-red-600 mt-1">{{ commonDataError }}</div>
    </div>

    <div>
      <div class="text-sm text-gray-600 mb-2">전송 공통 headers (JSON, 선택)</div>
      <textarea v-model="commonHeadersText" rows="2" class="w-full px-3 py-2 border rounded" placeholder='{"X-A":"1"}'></textarea>
      <div v-if="commonHeadersError" class="text-xs text-red-600 mt-1">{{ commonHeadersError }}</div>
    </div>

    <!-- 액션 -->
    <div class="flex items-center justify-end gap-2 pt-2">
      <button @click="$emit('cancel')" type="button" class="px-4 py-2 rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50">취소</button>
      <button @click="submit" :disabled="submitting" class="px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50">
        {{ submitting ? '전송 중...' : '전송 시작' }}
      </button>
    </div>

    <!-- List picker modal -->
    <ListPickerModal v-model="listPickerOpen" :initialSelected="selectedLists" @confirmed="onListsConfirmed" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import ListPickerModal from './ListPickerModal.vue';
import { createCampaignDeliveries } from '../api';
import type { Campaign, List } from '../types';

// Props
const props = defineProps<{
  campaignId: string;
}>();

const emit = defineEmits<{
  (e: 'saved', payload: any): void;
  (e: 'cancel'): void;
}>();

// state
const selectedLists = ref<List[]>([]);
const listPickerOpen = ref(false);

const individuals = ref<
  Array<{
    _localId: string;
    email: string;
    name?: string;
    dataText: string;
    headersText: string;
    dataError?: string | null;
    headersError?: string | null;
  }>
>([]);

const commonDataText = ref<string>('');
const commonHeadersText = ref<string>('');
const commonDataError = ref<string | null>(null);
const commonHeadersError = ref<string | null>(null);

const scheduledAtLocal = ref<string>(''); // datetime-local string
const submitting = ref(false);

// helpers
const openListPicker = () => {
  listPickerOpen.value = true;
};

const onListsConfirmed = (listsSelected: List[] | string[]) => {
  // ListPickerModal now emits full List objects; but keep fallback for string IDs
  if (listsSelected.length > 0 && typeof listsSelected[0] === 'object') {
    selectedLists.value = (listsSelected as List[]).slice();
  } else {
    // fallback: convert ids to minimal objects (name unknown)
    selectedLists.value = (listsSelected as string[]).map((id: string) => ({ id, name: id } as List));
  }
};

const removeList = (id: string) => {
  selectedLists.value = selectedLists.value.filter((x: List) => x.id !== id);
};

const addIndividual = () => {
  individuals.value.push({
    _localId: String(Math.random()).slice(2),
    email: '',
    name: '',
    dataText: '{}',
    headersText: '{}',
    dataError: null,
    headersError: null,
  });
};

const removeIndividual = (idx: number) => {
  individuals.value.splice(idx, 1);
};

const clearIndividuals = () => {
  individuals.value = [];
};

const parseJsonSafe = (text: string) => {
  if (!text || text.trim() === '') return {};
  try {
    return JSON.parse(text);
  } catch (err: any) {
    throw new Error(err?.message || 'JSON 파싱 오류');
  }
};

const validateIndividuals = () => {
  let ok = true;
  for (const ind of individuals.value) {
    ind.dataError = null;
    ind.headersError = null;
    try {
      parseJsonSafe(ind.dataText);
    } catch (e: any) {
      ind.dataError = 'data JSON이 유효하지 않습니다: ' + e.message;
      ok = false;
    }
    try {
      parseJsonSafe(ind.headersText);
    } catch (e: any) {
      ind.headersError = 'headers JSON이 유효하지 않습니다: ' + e.message;
      ok = false;
    }
    if (!ind.email || ind.email.trim() === '') {
      ind.dataError = ind.dataError || '이메일을 입력하세요';
      ok = false;
    }
  }
  return ok;
};

const validateCommon = () => {
  commonDataError.value = null;
  commonHeadersError.value = null;
  try {
    parseJsonSafe(commonDataText.value || '');
  } catch (e: any) {
    commonDataError.value = '공통 data JSON이 유효하지 않습니다: ' + e.message;
    return false;
  }
  try {
    parseJsonSafe(commonHeadersText.value || '');
  } catch (e: any) {
    commonHeadersError.value = '공통 headers JSON이 유효하지 않습니다: ' + e.message;
    return false;
  }
  return true;
};

const toUnixSeconds = (dtLocal: string) => {
  if (!dtLocal) return undefined;
  const d = new Date(dtLocal);
  if (isNaN(d.getTime())) return undefined;
  return Math.floor(d.getTime() / 1000);
};

const submit = async () => {
  // at least one target
  if (selectedLists.value.length === 0 && individuals.value.length === 0) {
    alert('리스트 또는 개별 수신자 중 하나 이상을 지정해야 합니다.');
    return;
  }

  if (!validateCommon()) return;
  if (!validateIndividuals()) return;

  let commonDataObj = {};
  let commonHeadersObj = {};
  try {
    commonDataObj = parseJsonSafe(commonDataText.value || '');
    commonHeadersObj = parseJsonSafe(commonHeadersText.value || '');
  } catch {
    // already handled in validateCommon
    return;
  }

  const individualsPayload = individuals.value.map((ind: any) => {
    const dataObj = parseJsonSafe(ind.dataText || '{}');
    const headersObj = parseJsonSafe(ind.headersText || '{}');
    return {
      email: ind.email,
      name: ind.name || undefined,
      data: { ...commonDataObj, ...dataObj },
      headers: { ...commonHeadersObj, ...headersObj },
    };
  });

  const body: any = {};
  if (selectedLists.value.length > 0) body.lists = selectedLists.value.map((l: List) => l.id);
  if (individualsPayload.length > 0) body.individuals = individualsPayload;
  const scheduled = toUnixSeconds(scheduledAtLocal.value);
  if (scheduled) body.scheduled_at = scheduled;

  submitting.value = true;
  try {
    const resp = await createCampaignDeliveries(props.campaignId, body);
    emit('saved', resp);
    alert('전송 요청이 접수되었습니다.');
    // reset form
    selectedLists.value = [];
    individuals.value = [];
    commonDataText.value = '';
    commonHeadersText.value = '';
    scheduledAtLocal.value = '';
  } catch (err) {
    console.error('전송 실패', err);
    alert('전송 요청에 실패했습니다.');
  } finally {
    submitting.value = false;
  }
};
</script>
