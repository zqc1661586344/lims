import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录', noAuth: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '工作台', icon: 'Odometer' },
      },
      {
        path: 'system/users',
        name: 'SystemUsers',
        component: () => import('@/views/system/user/index.vue'),
        meta: { title: '用户管理', icon: 'User' },
      },
      {
        path: 'system/depts',
        name: 'SystemDepts',
        component: () => import('@/views/system/dept/index.vue'),
        meta: { title: '部门管理', icon: 'Organization' },
      },
      {
        path: 'system/roles',
        name: 'SystemRoles',
        component: () => import('@/views/system/role/index.vue'),
        meta: { title: '角色管理', icon: 'Key' },
      },
      // Base Data
      {
        path: 'base-data/items',
        name: 'BaseDataItems',
        component: () => import('@/views/baseData/items/index.vue'),
        meta: { title: '检测项目', icon: 'List' },
      },
      {
        path: 'base-data/standards',
        name: 'BaseDataStandards',
        component: () => import('@/views/baseData/standards/index.vue'),
        meta: { title: '检测标准', icon: 'Document' },
      },
      {
        path: 'base-data/equipment',
        name: 'BaseDataEquipment',
        component: () => import('@/views/baseData/equipment/index.vue'),
        meta: { title: '仪器设备', icon: 'Monitor' },
      },
      {
        path: 'base-data/reagents',
        name: 'BaseDataReagents',
        component: () => import('@/views/baseData/reagents/index.vue'),
        meta: { title: '试剂耗材', icon: 'Box' },
      },
      // Business
      {
        path: 'business/pending-tasks',
        name: 'PendingTasks',
        component: () => import('@/views/business/pendingTasks/index.vue'),
        meta: { title: '待办任务', icon: 'Bell' },
      },
      {
        path: 'business/task-order',
        name: 'TaskOrder',
        component: () => import('@/views/business/taskOrder/index.vue'),
        meta: { title: '任务委托', icon: 'Edit' },
      },
      {
        path: 'business/contract-review',
        name: 'ContractReview',
        component: () => import('@/views/business/contractReview/index.vue'),
        meta: { title: '合同评审', icon: 'Document' },
      },
      {
        path: 'business/qc-task',
        name: 'QCTask',
        component: () => import('@/views/business/qcTask/index.vue'),
        meta: { title: '质控任务', icon: 'DataAnalysis' },
      },
      {
        path: 'business/sampling-schedule',
        name: 'SamplingSchedule',
        component: () => import('@/views/business/samplingSchedule/index.vue'),
        meta: { title: '采样调度', icon: 'Calendar' },
      },
      {
        path: 'business/field-sampling',
        name: 'FieldSampling',
        component: () => import('@/views/business/fieldSampling/index.vue'),
        meta: { title: '现场采样', icon: 'Location' },
      },
      {
        path: 'business/sample-receiving',
        name: 'SampleReceiving',
        component: () => import('@/views/business/sampleReceiving/index.vue'),
        meta: { title: '样品接收', icon: 'Box' },
      },
      {
        path: 'business/task-assign',
        name: 'TaskAssign',
        component: () => import('@/views/business/taskAssign/index.vue'),
        meta: { title: '任务分配', icon: 'UserFilled' },
      },
      {
        path: 'business/data-entry',
        name: 'DataEntry',
        component: () => import('@/views/business/dataEntry/index.vue'),
        meta: { title: '数据录入', icon: 'EditPen' },
      },
      {
        path: 'business/data-review',
        name: 'DataReview',
        component: () => import('@/views/business/dataReview/index.vue'),
        meta: { title: '数据复核', icon: 'Search' },
      },
      {
        path: 'business/data-audit',
        name: 'DataAudit',
        component: () => import('@/views/business/dataAudit/index.vue'),
        meta: { title: '数据审核', icon: 'Finished' },
      },
      {
        path: 'business/report-prepare',
        name: 'ReportPrepare',
        component: () => import('@/views/business/reportPrepare/index.vue'),
        meta: { title: '报告编制', icon: 'Document' },
      },
      {
        path: 'business/report-review',
        name: 'ReportReview',
        component: () => import('@/views/business/reportReview/index.vue'),
        meta: { title: '报告复核', icon: 'Search' },
      },
      {
        path: 'business/report-audit',
        name: 'ReportAudit',
        component: () => import('@/views/business/reportAudit/index.vue'),
        meta: { title: '报告审核', icon: 'Finished' },
      },
      {
        path: 'business/report-sign',
        name: 'ReportSign',
        component: () => import('@/views/business/reportSign/index.vue'),
        meta: { title: '报告签发', icon: 'Stamp' },
      },
      {
        path: 'business/report-print',
        name: 'ReportPrint',
        component: () => import('@/views/business/reportPrint/index.vue'),
        meta: { title: '报告打印发放', icon: 'Printer' },
      },
      {
        path: 'business/project-archive',
        name: 'ProjectArchive',
        component: () => import('@/views/business/projectArchive/index.vue'),
        meta: { title: '项目归档', icon: 'FolderOpened' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard — will be enhanced in Phase 2 with auth check
router.beforeEach((to, _from, next) => {
  const token = sessionStorage.getItem('token') || localStorage.getItem('token')
  if (to.meta.noAuth) {
    next()
  } else if (!token) {
    next('/login')
  } else {
    next()
  }
})

export default router