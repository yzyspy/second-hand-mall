import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AdminLayout from '@/layouts/AdminLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/pages/Login.vue') },
    {
      path: '/',
      component: AdminLayout,
      redirect: '/users',
      children: [
        { path: 'users',        component: () => import('@/pages/users/UserList.vue') },
        { path: 'users/:id',    component: () => import('@/pages/users/UserDetail.vue') },
        { path: 'products',     component: () => import('@/pages/products/ProductList.vue') },
        { path: 'products/:id', component: () => import('@/pages/products/ProductDetail.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/users' },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.path !== '/login' && !auth.token) {
    return '/login'
  }
})

export default router
