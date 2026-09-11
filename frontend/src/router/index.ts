import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/samples',
    },
    {
      path: '/samples',
      name: 'samples',
      component: () => import('@/views/SampleListView.vue'),
    },
    {
      path: '/samples/:id',
      name: 'sample-detail',
      component: () => import('@/views/SampleDetailView.vue'),
      props: (route) => ({ id: Number(route.params.id) }),
    },
  ],
})

export default router
