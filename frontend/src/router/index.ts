import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/lots',
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
    {
      path: '/lots',
      name: 'lots',
      component: () => import('@/views/LotListView.vue'),
    },
    {
      path: '/lots/:id',
      name: 'lot-detail',
      component: () => import('@/views/LotDetailView.vue'),
      props: (route) => ({ id: Number(route.params.id) }),
    },
    {
      path: '/machines',
      name: 'machines',
      component: () => import('@/views/MachineBoardView.vue'),
    },
  ],
})

export default router
