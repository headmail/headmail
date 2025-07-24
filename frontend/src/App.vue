<script setup lang="ts">
// No script needed here for now, as content is handled by router-view
</script>

<template>
  <div class="flex h-screen bg-gray-100">
    <!-- Sidebar -->
    <div class="w-64 bg-gray-800 text-white">
      <div class="p-4 text-lg font-bold">Headmail</div>
      <nav>
        <ul>
          <router-link to="/campaigns" v-slot="{ href, navigate, isActive }">
            <li :class="[isActive ? 'bg-gray-700' : '', 'p-4 hover:bg-gray-700 cursor-pointer']">
              <a :href="href" @click="navigate">Campaigns</a>
            </li>
          </router-link>
          <router-link to="/lists" v-slot="{ href, navigate, isActive }">
            <li :class="[isActive ? 'bg-gray-700' : '', 'p-4 hover:bg-gray-700 cursor-pointer']">
              <a :href="href" @click="navigate">Lists</a>
            </li>
          </router-link>
          <router-link to="/subscribers" v-slot="{ href, navigate, isActive }">
            <li :class="[isActive ? 'bg-gray-700' : '', 'p-4 hover:bg-gray-700 cursor-pointer']">
              <a :href="href" @click="navigate">Subscribers</a>
            </li>
          </router-link>
          <router-link to="/templates" v-slot="{ href, navigate, isActive }">
            <li :class="[isActive ? 'bg-gray-700' : '', 'p-4 hover:bg-gray-700 cursor-pointer']">
              <a :href="href" @click="navigate">Templates</a>
            </li>
          </router-link>
        </ul>
      </nav>
    </div>

    <!-- Main content -->
    <div class="flex-1 p-10 overflow-y-auto">
      <router-view />
    </div>
  </div>
</template>

<style scoped>
</style>
