<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h1 class="text-2xl font-bold">Lists</h1>
      <button @click="showCreateModal = true" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">Create List</button>
    </div>

    <!-- Search and filter controls -->
    <div class="mb-4">
      <input v-model="searchQuery" @input="fetchLists" type="text" placeholder="Search lists..." class="p-2 border rounded w-full">
    </div>

    <table class="min-w-full bg-white">
      <thead>
        <tr>
          <th class="py-2 px-4 border-b">Name</th>
          <th class="py-2 px-4 border-b">Description</th>
          <th class="py-2 px-4 border-b">Subscribers</th>
          <th class="py-2 px-4 border-b">Created At</th>
          <th class="py-2 px-4 border-b">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="list in lists" :key="list.id">
          <td class="py-2 px-4 border-b">{{ list.name }}</td>
          <td class="py-2 px-4 border-b">{{ list.description }}</td>
          <td class="py-2 px-4 border-b">{{ list.subscriber_count }}</td>
          <td class="py-2 px-4 border-b">{{ list.created_at ? new Date(list.created_at * 1000).toLocaleString() : '' }}</td>
          <td class="py-2 px-4 border-b">
            <button @click="editList(list)" class="text-blue-500 hover:underline mr-2">Edit</button>
            <button @click="confirmDelete(list)" class="text-red-500 hover:underline">Delete</button>
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
    <div v-if="showCreateModal || editingList" class="fixed inset-0 bg-gray-800 bg-opacity-50 flex items-center justify-center">
      <div class="bg-white p-6 rounded-lg shadow-lg w-1/3">
        <h2 class="text-xl font-bold mb-4">{{ editingList ? 'Edit List' : 'Create List' }}</h2>
        <form @submit.prevent="saveList">
          <div class="mb-4">
            <label for="name" class="block text-sm font-medium text-gray-700">Name</label>
            <input v-model="listForm.name" type="text" id="name" class="mt-1 p-2 border rounded w-full" required>
          </div>
          <div class="mb-4">
            <label for="description" class="block text-sm font-medium text-gray-700">Description</label>
            <textarea v-model="listForm.description" id="description" rows="3" class="mt-1 p-2 border rounded w-full"></textarea>
          </div>
          <div class="flex justify-end">
            <button type="button" @click="closeModal" class="bg-gray-300 text-gray-800 px-4 py-2 rounded mr-2">Cancel</button>
            <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded">{{ editingList ? 'Update' : 'Create' }}</button>
          </div>
        </form>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { getLists, createList, updateList, deleteList } from '../../api';
import type { List, PaginationResponse } from '../../types';
import Pagination from '../../components/Pagination.vue';

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
    if (editingList.value && editingList.value.id) {
      await updateList(editingList.value.id, listForm);
    } else {
      await createList(listForm);
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
  showCreateModal.value = true;
};

const confirmDelete = async (list: List) => {
  if (list.id && window.confirm(`Are you sure you want to delete the list "${list.name}"?`)) {
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
};

onMounted(fetchLists);
</script>
