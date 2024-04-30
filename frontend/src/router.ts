import { createRouter, createWebHashHistory } from 'vue-router'

export default createRouter({
    history: createWebHashHistory(),
    linkActiveClass: "active",
    routes: [
        {
            path: '/',
            component: () => import('./components/Project.vue'),
        },
        {
            path: '/turbine',
            component: () => import('./components/Turbine.vue'),
        },
        {
            path: '/analysis',
            component: () => import('./components/Analysis.vue'),
        },
        {
            path: '/evaluate',
            component: () => import('./components/Evaluate.vue'),
        },
        {
            path: '/results',
            component: () => import('./components/Results.vue'),
        },
        {
            path: '/vtk',
            component: () => import('./components/Vtk.vue'),
        },
    ],
})
