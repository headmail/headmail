<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">캠페인 상세</h1>
      <div class="flex items-center gap-2">
        <button :class="activeTab === 'detail' ? 'px-4 py-2 bg-blue-600 text-white rounded' : 'px-4 py-2 border rounded'" @click="activeTab = 'detail'">상세</button>
        <button :class="activeTab === 'stats' ? 'px-4 py-2 bg-blue-600 text-white rounded' : 'px-4 py-2 border rounded'" @click="activeTab = 'stats'">통계</button>
        <button :class="activeTab === 'send' ? 'px-4 py-2 bg-blue-600 text-white rounded' : 'px-4 py-2 border rounded'" @click="activeTab = 'send'">전송</button>
      </div>
    </div>

    <div v-if="loading" class="text-gray-600">로딩 중...</div>

    <div v-else>
      <div v-if="activeTab === 'detail'">
        <div class="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">이름</label>
            <input v-model="campaign.name" type="text" class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md" />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">제목</label>
            <input v-model="campaign.subject" type="text" class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md" :placeholder="subjectPlaceholder" />
          </div>

        <div>
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
          <textarea v-model="campaign.template_html" rows="10" class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"></textarea>

          <label class="block text-sm font-medium text-gray-700 mt-3">Plain Text</label>
          <textarea v-model="campaign.template_text" rows="6" class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"></textarea>
        </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-sm text-gray-500">ID</div>
              <div class="text-gray-900 font-medium">{{ campaign.id }}</div>
            </div>
            <div>
              <div class="text-sm text-gray-500">상태</div>
              <div class="text-gray-900 font-medium">{{ campaign.status }}</div>
            </div>
          </div>

          <div class="flex justify-end">
            <button @click="handleSave" class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">저장</button>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'stats'">
        <StatsChart :campaignId="campaign.id!!" />
      </div>

      <div v-if="activeTab === 'send'">
        <DeliveryForm :campaignId="campaign.id!!" @saved="onDeliverySaved" @cancel="onDeliveryCancel" />
      </div>
    </div>

    <TemplatePickerModal
        :modelValue="pickerVisible"
        :initialSelected="campaign ? campaign.template_id : null"
        @update:modelValue="pickerVisible = $event"
        @confirmed="onTemplateConfirmed"
    />
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, computed} from 'vue';
import { useRoute } from 'vue-router';
import { getCampaign, updateCampaign, getTemplate } from '../../api';
import DeliveryForm from '../../components/DeliveryForm.vue';
import StatsChart from '../../components/StatsChart.vue';
import TemplatePickerModal from '../../components/TemplatePickerModal.vue';
import type { Campaign } from '../../types';

const route = useRoute();
const campaign = ref<Campaign>({
  id: '',
  name: '',
  subject: '',
  status: '',
  template_id: '',
  template_html: '',
  template_text: '',
});
const loading = ref(true);
const activeTab = ref<'detail' | 'send' | 'stats'>('detail');

const pickerVisible = ref(false);
const selectedTemplateName = ref<string | null>(null);

const subjectPlaceholder = computed(() => {
  if (!campaign.value.template_id) {
    return '';
  } else {
    return '템플릿의 제목으로 대체됩니다';
  }
});

const fetchCampaign = async () => {
  loading.value = true;
  const campaignId = route.params.id as string;
  try {
    const resp = await getCampaign(campaignId);
    campaign.value = {
      ...resp,
      template_id: (resp as any).template_id || (resp as any).templateID || null,
      template_html: (resp as any).template_html || (resp as any).templateHTML || '',
      template_text: (resp as any).template_text || (resp as any).templateText || '',
    };
    if (campaign.value.template_id) {
      await loadTemplateName(campaign.value.template_id);
    } else {
      selectedTemplateName.value = null;
    }
  } catch (err) {
    console.error('캠페인 로드 실패', err);
    alert('캠페인 로드에 실패했습니다.');
  } finally {
    loading.value = false;
  }
};

onMounted(fetchCampaign);

const openPicker = () => {
  pickerVisible.value = true;
};

const clearTemplate = () => {
  campaign.value.template_id = '';
  selectedTemplateName.value = null;
};

const onTemplateConfirmed = async (templateID: string | null) => {
  pickerVisible.value = false;
  campaign.value.template_id = '';
  campaign.value.subject = '';
  if (templateID) {
    await loadTemplateName(templateID);
    campaign.value.template_html = '';
    campaign.value.template_text = '';
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

const handleSave = async () => {
  if (!campaign.value || !campaign.value.id) return;
  try {
    await updateCampaign(campaign.value.id, {
      name: campaign.value.name,
      subject: campaign.value.subject,
      template_id: campaign.value.template_id,
      template_html: campaign.value.template_html,
      template_text: campaign.value.template_text,
    } as any);
    alert('저장되었습니다.');
    await fetchCampaign();
  } catch (err) {
    console.error('저장 실패', err);
    alert('저장에 실패했습니다.');
  }
};

const onDeliverySaved = (resp: any) => {
  alert('전송 요청이 접수되었습니다.');
  fetchCampaign();
};

const onDeliveryCancel = () => {
  activeTab.value = 'detail';
};
</script>
