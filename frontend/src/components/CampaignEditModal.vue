<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center">
    <div class="absolute inset-0 bg-black/50" @click="close"></div>
    <div class="relative bg-white w-full max-w-lg mx-4 rounded-2xl shadow-xl border border-gray-200 p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold text-gray-900">캠페인 수정</h2>
        <button @click="close" class="p-2 text-gray-400 hover:text-gray-600">
          <span class="sr-only">닫기</span>
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <div v-if="form" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700">이름</label>
          <input v-model="form.name" type="text" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700">제목</label>
          <input v-model="form.subject" type="text" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button @click="close" type="button" class="px-4 py-2 rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50">
            취소
          </button>
          <button @click="save" :disabled="saving" class="px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50">
            {{ saving ? '저장 중...' : '저장' }}
          </button>
        </div>
      </div>

      <div v-else class="text-gray-600">로딩중...</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, toRaw } from 'vue';
import type { Campaign } from '../types';
import { updateCampaign } from '../api';

const props = defineProps<{
  modelValue: boolean;
  campaign: Campaign | null;
}>();

const emit = defineEmits(['update:modelValue', 'saved']);

const visible = ref<boolean>(props.modelValue || false);
const saving = ref(false);
const form = ref<{ id: string | number; name: string; subject: string } | null>(null);

watch(
  () => props.modelValue,
  (v) => {
    visible.value = v;
    if (!v) {
      form.value = null;
    }
  }
);

watch(
  () => props.campaign,
  (c) => {
    if (c) {
      form.value = {
        id: c.id as any,
        name: c.name || '',
        subject: c.subject || '',
      };
    } else {
      form.value = null;
    }
  },
  { immediate: true }
);

const close = () => {
  emit('update:modelValue', false);
};

const save = async () => {
  if (!form.value) return;
  saving.value = true;
  try {
    const updated = await updateCampaign(String(form.value.id), {
      name: form.value.name,
      subject: form.value.subject,
    } as any);
    emit('saved', updated);
    close();
  } catch (err) {
    console.error('캠페인 업데이트 실패:', err);
    alert('캠페인 업데이트에 실패했습니다.');
  } finally {
    saving.value = false;
  }
};
</script>
