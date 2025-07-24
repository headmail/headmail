<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h1 class="text-2xl font-bold">Subscribers</h1>
      <!-- Add subscriber button will be implemented later, as it requires selecting a list -->
    </div>

    <!-- Search and filter controls -->
    <div class="mb-4">
      <input v-model="searchQuery" @input="fetchSubscribers" type="text" placeholder="Search subscribers..." class="p-2 border rounded w-full">
    </div>

    <table class="min-w-full bg-white">
      <thead>
        <tr>
          <th class="py-2 px-4 border-b">Name</th>
          <th class="py-2 px-4 border-b">Email</th>
          <th class="py-2 px-4 border-b">Status</th>
          <th class="py-2 px-4 border-b">Created At</th>
          <th class="py-2 px-4 border-b">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="subscriber in subscribers" :key="subscriber.id">
          <td class="py-2 px-4 border-b">{{ subscriber.name }}</td>
          <td class="py-2 px-4 border-b">{{ subscriber.email }}</td>
          <td class="py-2 px-4 border-b">{{ subscriber.status }}</td>
          <td class="py-2 px-4 border-b">{{ subscriber.created_at ? new Date(subscriber.created_at * 1000).toLocaleString() : '' }}</td>
          <td class="py-2 px-4 border-b">
            <button @click="editSubscriber(subscriber)" class="text-blue-500 hover:underline mr-2">Edit</button>
            <button @click="confirmDelete(subscriber)" class="text-red-500 hover:underline">Delete</button>
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

    <!-- Edit Modal -->
    <div v-if="editingSubscriber" class="fixed inset-0 bg-gray-800 bg-opacity-50 flex items-center justify-center">
      <div class="bg-white p-6 rounded-lg shadow-lg w-1/3">
        <h2 class="text-xl font-bold mb-4">Edit Subscriber</h2>
        <form @submit.prevent="saveSubscriber">
          <div class="mb-4">
            <label for="name" class="block text-sm font-medium text-gray-700">Name</label>
            <input v-model="subscriberForm.name" type="text" id="name" class="mt-1 p-2 border rounded w-full" required>
          </div>
          <div class="mb-4">
            <label for="email" class="block text-sm font-medium text-gray-700">Email</label>
            <input v-model="subscriberForm.email" type="email" id="email" class="mt-1 p-2 border rounded w-full" required>
          </div>
           <div class="mb-4">
            <label for="status" class="block text-sm font-medium text-gray-700">Status</label>
            <select v-model="subscriberForm.status" id="status" class="mt-1 p-2 border rounded w-full">
              <option value="enabled">Enabled</option>
              <option value="disabled">Disabled</option>
            </select>
          </div>
          <div class="flex justify-end">
            <button type="button" @click="closeModal" class="bg-gray-300 text-gray-800 px-4 py-2 rounded mr-2">Cancel</button>
            <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded">Update</button>
          </div>
        </form>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { getSubscribers, updateSubscriber, deleteSubscriber } from '../../api';
import type { Subscriber, PaginationResponse } from '../../types';
import Pagination from '../../components/Pagination.vue';

const subscribers = ref<Subscriber[]>([]);
const pagination = ref<PaginationResponse | null>(null);
const searchQuery = ref('');
const currentPage = ref(1);
const limit = ref(20);
const editingSubscriber = ref<Subscriber | null>(null);

const subscriberForm = reactive({
  name: '',
  email: '',
  status: 'enabled' as 'enabled' | 'disabled',
});

const fetchSubscribers = async () => {
  try {
    const response = await getSubscribers({ 
      page: currentPage.value,
      limit: limit.value,
      search: searchQuery.value || undefined 
    });
    
    if (response && typeof response === 'object' && 'data' in response) {
      const paginatedResponse = response as any;
      subscribers.value = paginatedResponse.data || [];
      pagination.value = paginatedResponse.pagination || null;
    }
  } catch (error) {
    console.error('Failed to fetch subscribers:', error);
  }
};

const saveSubscriber = async () => {
  if (!editingSubscriber.value || !editingSubscriber.value.id) return;
  try {
    await updateSubscriber(editingSubscriber.value.id, subscriberForm);
    fetchSubscribers();
    closeModal();
  } catch (error) {
    console.error('Failed to save subscriber:', error);
  }
};

const editSubscriber = (subscriber: Subscriber) => {
  editingSubscriber.value = subscriber;
  subscriberForm.name = subscriber.name || '';
  subscriberForm.email = subscriber.email || '';
  subscriberForm.status = subscriber.status || 'enabled';
};

const confirmDelete = async (subscriber: Subscriber) => {
  if (subscriber.id && window.confirm(`Are you sure you want to delete the subscriber "${subscriber.name}"?`)) {
    try {
      await deleteSubscriber(subscriber.id);
      fetchSubscribers();
    } catch (error) {
      console.error('Failed to delete subscriber:', error);
    }
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchSubscribers();
};

const closeModal = () => {
  editingSubscriber.value = null;
  subscriberForm.name = '';
  subscriberForm.email = '';
  subscriberForm.status = 'enabled';
};

onMounted(fetchSubscribers);
</script>
