<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h1 class="text-2xl font-bold">Templates</h1>
      <button @click="showCreateModal = true" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">Create Template</button>
    </div>

    <table class="min-w-full bg-white">
      <thead>
        <tr>
          <th class="py-2 px-4 border-b">Name</th>
          <th class="py-2 px-4 border-b">Created At</th>
          <th class="py-2 px-4 border-b">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="template in templates" :key="template.id">
          <td class="py-2 px-4 border-b">{{ template.name }}</td>
          <td class="py-2 px-4 border-b">{{ template.created_at ? new Date(template.created_at * 1000).toLocaleString() : '' }}</td>
          <td class="py-2 px-4 border-b">
            <button @click="editTemplate(template)" class="text-blue-500 hover:underline mr-2">Edit</button>
            <button @click="confirmDelete(template)" class="text-red-500 hover:underline">Delete</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Pagination -->
    <Pagination
      v-if="pagination"
      :current-page="pagination.page || 1"
      :total="pagination.total || 0"
      :limit="pagination.limit || 20"
      @page-change="handlePageChange"
    />

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || editingTemplate" class="fixed inset-0 bg-gray-800 bg-opacity-50 flex items-center justify-center">
      <div class="bg-white p-6 rounded-lg shadow-lg w-2/3">
        <h2 class="text-xl font-bold mb-4">{{ editingTemplate ? 'Edit Template' : 'Create Template' }}</h2>
        <form @submit.prevent="saveTemplate">
          <div class="mb-4">
            <label for="name" class="block text-sm font-medium text-gray-700">Name</label>
            <input v-model="templateForm.name" type="text" id="name" class="mt-1 p-2 border rounded w-full" required>
          </div>
          <div class="mb-4">
            <label for="body_html" class="block text-sm font-medium text-gray-700">HTML Body</label>
            <textarea v-model="templateForm.body_html" id="body_html" rows="10" class="mt-1 p-2 border rounded w-full"></textarea>
          </div>
          <div class="mb-4">
            <label for="body_text" class="block text-sm font-medium text-gray-700">Text Body</label>
            <textarea v-model="templateForm.body_text" id="body_text" rows="10" class="mt-1 p-2 border rounded w-full"></textarea>
          </div>
          <div class="flex justify-end">
            <button type="button" @click="closeModal" class="bg-gray-300 text-gray-800 px-4 py-2 rounded mr-2">Cancel</button>
            <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded">{{ editingTemplate ? 'Update' : 'Create' }}</button>
          </div>
        </form>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { getTemplates, createTemplate, updateTemplate, deleteTemplate } from '../../api';
import type { Template, PaginationResponse } from '../../types';
import Pagination from '../../components/Pagination.vue';

const templates = ref<Template[]>([]);
const pagination = ref<PaginationResponse | null>(null);
const currentPage = ref(1);
const limit = ref(20);
const showCreateModal = ref(false);
const editingTemplate = ref<Template | null>(null);

const templateForm = reactive({
  name: '',
  body_html: '',
  body_text: '',
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
  templateForm.body_html = template.body_html || '';
  templateForm.body_text = template.body_text || '';
  showCreateModal.value = true;
};

const confirmDelete = async (template: Template) => {
  if (template.id && window.confirm(`Are you sure you want to delete the template "${template.name}"?`)) {
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
  templateForm.body_html = '';
  templateForm.body_text = '';
};

onMounted(fetchTemplates);
</script>
