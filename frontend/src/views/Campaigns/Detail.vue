<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">캠페인 상세</h1>
      <div class="flex items-center gap-2">
        <button :class="activeTab === 'detail' ? 'px-4 py-2 bg-blue-600 text-white rounded' : 'px-4 py-2 border rounded'" @click="activeTab = 'detail'">상세</button>
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
            <input v-model="campaign.subject" type="text" class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md" />
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

      <div v-if="activeTab === 'send'">
        <DeliveryForm :campaignId="campaign.id" @saved="onDeliverySaved" @cancel="onDeliveryCancel" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { getCampaign, updateCampaign } from '../../api';
import DeliveryForm from '../../components/DeliveryForm.vue';
import type { Campaign } from '../../types';

const route = useRoute();
const campaign = ref<Campaign>({
  id: '',
  name: '',
  subject: '',
  status: '',
} as Campaign);
const loading = ref(true);
const activeTab = ref<'detail' | 'send'>('detail');

const fetchCampaign = async () => {
  loading.value = true;
  const campaignId = route.params.id as string;
  try {
    const resp = await getCampaign(campaignId);
    campaign.value = resp;
  } catch (err) {
    console.error('캠페인 로드 실패', err);
    alert('캠페인 로드에 실패했습니다.');
  } finally {
    loading.value = false;
  }
};

onMounted(fetchCampaign);

const handleSave = async () => {
  if (!campaign.value || !campaign.value.id) return;
  try {
    await updateCampaign(campaign.value.id, {
      name: campaign.value.name,
      subject: campaign.value.subject,
    } as any);
    alert('저장되었습니다.');
    await fetchCampaign();
  } catch (err) {
    console.error('저장 실패', err);
    alert('저장에 실패했습니다.');
  }
};

const onDeliverySaved = (resp: any) => {
  // 사용자 요청에 따라 전송 후 목록 자동 표시를 하지 않음 — 단순 알림만 표시
  alert('전송 요청이 접수되었습니다.');
  // 필요하면 캠페인 재조회
  fetchCampaign();
  // 탭을 상세로 되돌릴지 여부는 남겨둠(현재는 그대로 둠)
};

const onDeliveryCancel = () => {
  // 아무 동작 없이 상세 탭으로 전환
  activeTab.value = 'detail';
};
</script>
