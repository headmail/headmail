<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div class="max-w-4xl mx-auto p-6">
    <h1 class="text-2xl font-semibold mb-4">새 캠페인 생성</h1>

    <div class="mb-4">
      <label class="block text-sm font-medium text-gray-700">이름</label>
      <input v-model="campaign.name" type="text" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
    </div>

    <div class="mb-4">
      <label class="block text-sm font-medium text-gray-700">제목</label>
      <input v-model="campaign.subject" type="text" :placeholder="subjectPlaceholder" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
    </div>

    <div class="mb-4">
      <label class="block text-sm font-medium text-gray-700">템플릿</label>
      <div class="flex items-center gap-2 mt-1">
        <div class="flex-1">
          <div class="text-sm text-gray-700" v-if="selectedTemplateName">{{ selectedTemplateName }}</div>
          <div class="text-sm text-gray-500" v-else>템플릿을 선택하거나 HTML/Text를 직접 입력하세요.</div>
        </div>
        <button @click="openPicker" type="button" class="px-3 py-2 border rounded bg-white hover:bg-gray-50">템플릿 선택</button>
        <button @click="clearTemplate" type="button" class="px-3 py-2 border rounded bg-white hover:bg-gray-50">선택 해제</button>
      </div>
    </div>

    <div v-if="!campaign.template_id" class="mb-4">
      <label class="block text-sm font-medium text-gray-700">HTML 템플릿</label>
      <textarea v-model="campaign.template_mjml" rows="10" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm"></textarea>

      <label class="block text-sm font-medium text-gray-700 mt-3">Plain Text</label>
      <textarea v-model="campaign.template_text" rows="6" class="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm"></textarea>
    </div>

    <div class="flex items-center justify-end gap-2">
      <button @click="handleCreate" :disabled="creating" class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-700 disabled:opacity-50">
        {{ creating ? '생성중...' : '캠페인 생성' }}
      </button>
    </div>

    <TemplatePickerModal
      :modelValue="pickerVisible"
      :initialSelected="campaign.template_id"
      @update:modelValue="pickerVisible = $event"
      @confirmed="onTemplateConfirmed"
    />
  </div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue';
import { useRouter } from 'vue-router';
import { createCampaign, getTemplate } from '../../api';
import TemplatePickerModal from '../../components/TemplatePickerModal.vue';

const router = useRouter();

const campaign = ref<any>({
  name: '',
  subject: '',
  template_id: null,
  template_mjml: '',
  template_text: '',
});

const creating = ref(false);
const pickerVisible = ref(false);
const selectedTemplateName = ref<string | null>(null);

const subjectPlaceholder = computed(() => {
  if (!campaign.value.template_id) {
    return '';
  } else {
    return '템플릿의 제목으로 대체됩니다';
  }
});

const openPicker = () => {
  pickerVisible.value = true;
};

const clearTemplate = () => {
  campaign.value.template_id = null;
  selectedTemplateName.value = null;
};

const onTemplateConfirmed = async (templateID: string | null) => {
  pickerVisible.value = false;
  campaign.value.template_id = templateID;
  if (templateID) {
    try {
      const tmpl = await getTemplate(templateID);
      selectedTemplateName.value = tmpl?.name || tmpl?.id || templateID;
      // clear inline content to rely on template at send-time
      campaign.value.template_mjml = '';
      campaign.value.template_text = '';
    } catch (err) {
      console.error('템플릿 로드 실패', err);
      selectedTemplateName.value = templateID;
    }
  } else {
    selectedTemplateName.value = null;
  }
};

const handleCreate = async () => {
  creating.value = true;
  try {
    const payload: any = {
      name: campaign.value.name,
      subject: campaign.value.subject,
      template_id: campaign.value.template_id,
      template_mjml: campaign.value.template_mjml,
      template_text: campaign.value.template_text,
    };
    await createCampaign(payload);
    router.push({ name: 'Campaigns' });
  } catch (err) {
    console.error('Failed to create campaign:', err);
    alert('캠페인 생성에 실패했습니다.');
  } finally {
    creating.value = false;
  }
};
</script>
