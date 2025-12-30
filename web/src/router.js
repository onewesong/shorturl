import { createRouter, createWebHistory } from "vue-router";
import Login from "./views/Login.vue";
import Links from "./views/Links.vue";
import LinkForm from "./views/LinkForm.vue";
import { apiMe } from "./api";

const router = createRouter({
  history: createWebHistory("/admin/"),
  routes: [
    { path: "/login", name: "login", component: Login },
    { path: "/", redirect: "/links" },
    { path: "/links", name: "links", component: Links, meta: { requiresAuth: true } },
    { path: "/links/new", name: "linkNew", component: LinkForm, meta: { requiresAuth: true } },
    { path: "/links/:id/edit", name: "linkEdit", component: LinkForm, meta: { requiresAuth: true } },
  ],
});

let authCached = false;
let authOk = false;

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return true;
  if (authCached && authOk) return true;

  try {
    await apiMe();
    authCached = true;
    authOk = true;
    return true;
  } catch (e) {
    authCached = true;
    authOk = false;
    return { name: "login", query: { next: to.fullPath } };
  }
});

export function resetAuthCache() {
  authCached = false;
  authOk = false;
}

export default router;

