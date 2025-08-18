<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center">
    <div class="absolute inset-0 bg-black/50" @click="close"></div>
    <div class="relative bg-white w-full max-w-2xl mx-4 rounded-2xl shadow-xl border border-gray-200 p-6">
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

        <div>
          <label class="block text-sm font-medium text-gray-700">템플릿</label>
          <div class="flex items-center gap-2 mt-1">
            <div class="flex-1">
              <div class="text-sm text-gray-700" v-if="selectedTemplateName">{{ selectedTemplateName }}</div>
              <div class="text-sm text-gray-500" v-else>직접 입력 또는 템플릿 미선택</div>
            </div>
            <button @click="openPicker" type="button" class="px-3 py-2 border rounded bg-white hover:bg-gray-50">템플릿 선택</button>
            <button @click="clearTemplate" type="button" class="px-3 py-2 border rounded bg-white hover:bg-gray-50">선택 해제</button>
          </div>
        </div>

        <div v-if="!form.template_id">
          <label class="block text-sm font-medium text-gray-700">MJML 템플릿</label>
          <textarea v-model="form.template_mjml" rows="10" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg"></textarea>
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

    <TemplatePickerModal
      :modelValue="pickerVisible"
      :initialSelected="form ? form.template_id : null"
      @update:modelValue="pickerVisible = $event"
      @confirmed="onTemplateConfirmed"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue';
import type { Campaign, Template } from '../types';
import { updateCampaign, getTemplates, getTemplate } from '../api';
import TemplatePickerModal from './TemplatePickerModal.vue';
import type {components} from "@/generated/api-types.ts";

const props = defineProps<{
  modelValue: boolean;
  campaign: Campaign | null;
}>();

const emit = defineEmits(['update:modelValue', 'saved']);

const visible = ref<boolean>(props.modelValue || false);
const saving = ref(false);

type FormType = {
  id: string | number | null;
  name: string;
  subject: string;
  template_id: string | null;
  template_mjml: string;
};

const form = ref<FormType | null>(null);
const pickerVisible = ref(false);
const selectedTemplateName = ref<string | null>(null);

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
        id: c.id as any || null,
        name: c.name || '',
        subject: c.subject || '',
        template_id: (c as any).template_id,
        template_mjml: (c as any).template_mjml,
      };
      // load template name if template_id present
      if (form.value.template_id) {
        loadTemplateName(form.value.template_id);
      } else {
        selectedTemplateName.value = null;
      }
    } else {
      form.value = null;
      selectedTemplateName.value = null;
    }
  },
  { immediate: true }
);

const openPicker = () => {
  pickerVisible.value = true;
};

const clearTemplate = () => {
  if (!form.value) return;
  form.value.template_id = null;
  selectedTemplateName.value = null;
};

const onTemplateConfirmed = async (templateID: string | null) => {
  pickerVisible.value = false;
  if (!form.value) return;
  form.value.template_id = templateID;
  if (templateID) {
    await loadTemplateName(templateID);
    // clear inline html/text so server/template precedence is used
    form.value.template_mjml = '';
  } else {
    selectedTemplateName.value = null;
  }
};

const loadTemplateName = async (templateID: string) => {
  try {
    const tmpl = await getTemplate(templateID);
    selectedTemplateName.value = tmpl?.name || tmpl?.id || templateID;
  } catch (err) {
    console.error('템플릿 이름 로드 실패', err);
    selectedTemplateName.value = templateID;
  }
};

const close = () => {
  emit('update:modelValue', false);
};

const save = async () => {
  if (!form.value) return;
  saving.value = true;
  try {
    const payload: components["schemas"]["github_com_headmail_headmail_pkg_api_admin_dto.UpdateCampaignRequest"] = {
      name: form.value.name,
      subject: form.value.subject,
      template_id: form.value.template_id || undefined,
      template_mjml: form.value.template_mjml,
    };
    const updated = await updateCampaign(String(form.value.id), payload);
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
