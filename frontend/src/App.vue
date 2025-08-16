<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<script setup lang="ts">
// No script needed here for now, as content is handled by router-view
</script>

<template>
  <div class="flex h-screen bg-gradient-to-br from-slate-50 to-blue-50">
    <!-- Sidebar -->
    <div class="w-72 bg-white shadow-xl border-r border-gray-200 flex flex-col">
      <!-- Logo/Header -->
      <div class="p-6 border-b border-gray-100">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-gradient-to-r from-blue-500 to-purple-600 rounded-xl flex items-center justify-center">
            <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 4.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>
            </svg>
          </div>
          <div>
            <h1 class="text-xl font-bold text-gray-900">Headmail</h1>
            <p class="text-sm text-gray-500">Email Marketing</p>
          </div>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 p-4 space-y-2">
        <router-link to="/campaigns" v-slot="{ href, navigate, isActive }">
          <a :href="href" @click="navigate" 
             :class="[
               isActive 
                 ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg' 
                 : 'text-gray-700 hover:bg-gray-100', 
               'flex items-center space-x-3 px-4 py-3 rounded-xl transition-all duration-200 group'
             ]">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
            </svg>
            <span class="font-medium">캠페인</span>
          </a>
        </router-link>

        <router-link to="/lists" v-slot="{ href, navigate, isActive }">
          <a :href="href" @click="navigate" 
             :class="[
               isActive 
                 ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg' 
                 : 'text-gray-700 hover:bg-gray-100', 
               'flex items-center space-x-3 px-4 py-3 rounded-xl transition-all duration-200 group'
             ]">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
            </svg>
            <span class="font-medium">리스트</span>
          </a>
        </router-link>

        <router-link to="/subscribers" v-slot="{ href, navigate, isActive }">
          <a :href="href" @click="navigate" 
             :class="[
               isActive 
                 ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg' 
                 : 'text-gray-700 hover:bg-gray-100', 
               'flex items-center space-x-3 px-4 py-3 rounded-xl transition-all duration-200 group'
             ]">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"></path>
            </svg>
            <span class="font-medium">구독자</span>
          </a>
        </router-link>

        <router-link to="/templates" v-slot="{ href, navigate, isActive }">
          <a :href="href" @click="navigate" 
             :class="[
               isActive 
                 ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg' 
                 : 'text-gray-700 hover:bg-gray-100', 
               'flex items-center space-x-3 px-4 py-3 rounded-xl transition-all duration-200 group'
             ]">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
            </svg>
            <span class="font-medium">템플릿</span>
          </a>
        </router-link>
      </nav>

      <!-- Bottom section -->
      <div class="p-4 border-t border-gray-100">
        <div class="flex items-center space-x-3 px-4 py-3 bg-gray-50 rounded-xl">
          <div class="w-8 h-8 bg-gradient-to-r from-green-400 to-blue-500 rounded-full flex items-center justify-center">
            <span class="text-white text-sm font-medium">U</span>
          </div>
          <div class="flex-1">
            <p class="text-sm font-medium text-gray-900">사용자</p>
            <p class="text-xs text-gray-500">관리자</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Main content -->
    <div class="flex-1 overflow-hidden">
      <div class="h-full overflow-y-auto">
        <div class="p-8">
          <router-view />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
