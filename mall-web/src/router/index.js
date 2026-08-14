import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Detail from '../views/Detail.vue'
import Publish from '../views/Publish.vue'
import ChatList from '../views/ChatList.vue'
import My from '../views/My.vue'

const routes = [
  { path: '/', name: 'home', component: Home },
  { path: '/detail/:id', name: 'detail', component: Detail, props: true },
  { path: '/publish', name: 'publish', component: Publish },
  { path: '/chat', name: 'chat-list', component: ChatList },
  { path: '/my', name: 'my', component: My }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
