<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">이메일 템플릿</h1>
        <p class="text-gray-600 mt-1">이메일 캠페인에 사용할 템플릿을 관리하세요</p>
      </div>
      <button 
        @click="showCreateModal = true" 
        class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-600 text-white font-medium rounded-xl hover:from-blue-600 hover:to-purple-700 transition-all duration-200 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
        </svg>
        새 템플릿 만들기
      </button>
    </div>

    <!-- Templates Grid -->
    <div v-if="templates.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="template in templates" 
        :key="template.id"
        class="bg-white rounded-2xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-200 overflow-hidden group">
        
        <!-- Template Preview -->
        <div class="h-32 bg-gradient-to-br from-gray-50 to-gray-100 p-4 flex items-center justify-center border-b border-gray-100">
          <div class="text-center">
            <svg class="w-12 h-12 text-gray-400 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
            </svg>
            <p class="text-sm text-gray-500">이메일 템플릿</p>
          </div>
        </div>

        <!-- Template Info -->
        <div class="p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-2 truncate">{{ template.name }}</h3>
          <p class="text-sm text-gray-500 mb-4">
            {{ template.created_at ? new Date(template.created_at * 1000).toLocaleDateString('ko-KR') : '' }}
          </p>
          
          <!-- Actions -->
          <div class="flex items-center justify-between">
            <div class="flex space-x-2">
              <button 
                @click="editTemplate(template)"
                class="inline-flex items-center px-3 py-2 text-sm font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors duration-200">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                </svg>
                편집
              </button>
              <button 
                @click="confirmDelete(template)"
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
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-gray-900 mb-2">아직 템플릿이 없습니다</h3>
      <p class="text-gray-600 mb-6">첫 번째 이메일 템플릿을 만들어 시작해보세요</p>
      <button 
        @click="showCreateModal = true"
        class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-600 text-white font-medium rounded-xl hover:from-blue-600 hover:to-purple-700 transition-all duration-200">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
        </svg>
        새 템플릿 만들기
      </button>
    </div>

    <!-- Pagination -->
    <Pagination
      v-if="pagination && templates.length > 0"
      :current-page="pagination.page || 1"
      :total="pagination.total || 0"
      :limit="pagination.limit || 20"
      @page-change="handlePageChange"
    />

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || editingTemplate" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div :class="['bg-white shadow-2xl w-full overflow-hidden', fullscreen ? 'rounded-none max-w-none h-screen' : 'rounded-2xl max-w-6xl max-h-[90vh]']">
        <!-- Modal Header -->
        <div class="px-8 py-6 border-b border-gray-200 bg-gradient-to-r from-blue-50 to-purple-50">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-2xl font-bold text-gray-900">
                {{ editingTemplate ? '템플릿 편집' : '새 템플릿 만들기' }}
              </h2>
              <p class="text-gray-600 mt-1">이메일 캠페인에 사용할 템플릿을 작성하세요</p>
            </div>
            <div class="flex items-center">
              <button
                @click="fullscreen = !fullscreen"
                :title="fullscreen ? 'Exit fullscreen' : 'Fullscreen'"
                class="p-2 hover:bg-gray-100 rounded-full transition-colors duration-200 mr-2">
                <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 3H5a2 2 0 00-2 2v3m0 8v3a2 2 0 002 2h3M16 3h3a2 2 0 012 2v3M21 16v3a2 2 0 01-2 2h-3"></path>
                </svg>
              </button>
              <button 
                @click="closeModal"
                class="p-2 hover:bg-gray-100 rounded-full transition-colors duration-200">
                <svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Modal Body -->
        <div class="p-8 overflow-y-auto max-h-[calc(90vh-200px)]">
          <form @submit.prevent="saveTemplate" class="space-y-6">
            <!-- Template Name -->
            <div>
              <label for="name" class="block text-sm font-semibold text-gray-900 mb-2">
                템플릿 이름 <span class="text-red-500">*</span>
              </label>
              <input 
                v-model="templateForm.name" 
                type="text" 
                id="name" 
                class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                placeholder="예: 환영 이메일, 뉴스레터 등"
                required>
            </div>

            <!-- HTML / GrapesJS Body -->
            <div>
              <label for="body_html" class="block text-sm font-semibold text-gray-900 mb-2">
                콘텐츠
              </label>
              <div class="relative">
                <TemplateEditor
                  :modelValueMjml="templateForm.body_mjml"
                  :fullscreen="fullscreen"
                  :subject="templateForm.subject"
                  @update:mjml="(v) => templateForm.body_mjml = v"
                  @update:subject="(v) => templateForm.subject = v"
                />
              </div>
            </div>
          </form>
        </div>

        <!-- Modal Footer -->
        <div class="px-8 py-6 border-t border-gray-200 bg-gray-50 flex justify-end space-x-4">
          <button 
            type="button" 
            @click="closeModal" 
            class="px-6 py-3 border border-gray-300 text-gray-700 font-medium rounded-xl hover:bg-gray-100 transition-all duration-200">
            취소
          </button>
          <button 
            @click="saveTemplate"
            class="px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-600 text-white font-medium rounded-xl hover:from-blue-600 hover:to-purple-700 transition-all duration-200 shadow-lg">
            {{ editingTemplate ? '업데이트' : '생성' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { getTemplates, createTemplate, updateTemplate, deleteTemplate } from '../../api';
import type { Template, PaginationResponse } from '../../types';
import Pagination from '../../components/Pagination.vue';
import TemplateEditor from '../../components/TemplateEditor.vue';

const templates = ref<Template[]>([]);
const pagination = ref<PaginationResponse | null>(null);
const currentPage = ref(1);
const limit = ref(20);
const showCreateModal = ref(false);
const editingTemplate = ref<Template | null>(null);
const fullscreen = ref(false);

const templateForm = reactive({
  name: '',
  subject: '',
  body_mjml: '',
});

const fetchTemplates = async () => {
  try {
    const response = await getTemplates({
      page: currentPage.value,
      limit: limit.value
    });
    
    if (response && typeof response === 'object' && 'data' in response) {
      const paginatedResponse = response as any;
      templates.value = paginatedResponse.data || [];
      pagination.value = paginatedResponse.pagination || null;
    }
  } catch (error) {
    console.error('Failed to fetch templates:', error);
  }
};

const saveTemplate = async () => {
  try {
    if (editingTemplate.value && editingTemplate.value.id) {
      await updateTemplate(editingTemplate.value.id, templateForm);
    } else {
      await createTemplate(templateForm);
    }
    fetchTemplates();
    closeModal();
  } catch (error) {
    console.error('Failed to save template:', error);
  }
};

const editTemplate = (template: Template) => {
  editingTemplate.value = template;
  templateForm.name = template.name || '';
  templateForm.subject = (template as any).subject || '';
  templateForm.body_mjml = template.body_mjml || '';
  showCreateModal.value = true;
};

const confirmDelete = async (template: Template) => {
  if (template.id && window.confirm(`"${template.name}" 템플릿을 정말 삭제하시겠습니까?`)) {
    try {
      await deleteTemplate(template.id);
      fetchTemplates();
    } catch (error) {
      console.error('Failed to delete template:', error);
    }
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchTemplates();
};

const closeModal = () => {
  showCreateModal.value = false;
  editingTemplate.value = null;
  templateForm.name = '';
  templateForm.subject = '';
  templateForm.body_mjml = '';
  fullscreen.value = false;
};

onMounted(fetchTemplates);
</script>
