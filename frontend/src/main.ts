/**
 * Copyright 2025 JC-Lab
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router.ts';

createApp(App)
    .use(router)
    .mount('#app')
