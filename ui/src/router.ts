import { createRouter, createWebHistory } from 'vue-router'
import Browser from './components/Browser.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/s/local/'
    },
    {
      path: '/s/:storage/:path*',
      name: 'browse',
      component: Browser,
      props: (route) => ({
        storage: route.params.storage,
        path: route.params.path || '',
        snapshot: route.query.snapshot || null
      })
    }
  ]
})

export default router
