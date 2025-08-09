import { createRouter, createWebHistory } from 'vue-router';

const routes = [
  { path: '/', redirect: '/campaigns' },
  {
    path: '/campaigns',
    name: 'Campaigns',
    component: () => import('./views/Campaigns/Index.vue'),
  },
  {
    path: '/campaigns/create',
    name: 'CampaignCreate',
    component: () => import('./views/Campaigns/Create.vue'),
  },
  {
    path: '/campaigns/:id',
    name: 'CampaignDetail',
    component: () => import('./views/Campaigns/Detail.vue'),
  },
  {
    path: '/lists',
    name: 'Lists',
    component: () => import('./views/Lists/Index.vue'),
  },
  {
    path: '/subscribers',
    name: 'Subscribers',
    component: () => import('./views/Subscribers/Index.vue'),
  },
  {
    path: '/templates',
    name: 'Templates',
    component: () => import('./views/Templates/Index.vue'),
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
