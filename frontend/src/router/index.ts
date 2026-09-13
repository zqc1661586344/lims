import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

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
      ...(import.meta.env.DEV
        ? [
            {
              path: 'univer-test',
              name: 'UniverTest',
              component: () => import('@/views/test/UniverTest.vue'),
              meta: { title: 'Univer渲染测试', icon: 'DataAnalysis', noAuth: true },
            } as RouteRecordRaw,
          ]
        : []),
      {
        path: 'system/users',
        name: 'SystemUsers',
        component: () => import('@/views/system/user/index.vue'),
        meta: { title: '用户管理', icon: 'User', permission: 'system:users' },
      },
      {
        path: 'system/depts',
        name: 'SystemDepts',
        component: () => import('@/views/system/dept/index.vue'),
        meta: { title: '部门管理', icon: 'OfficeBuilding', permission: 'system:depts' },
      },
      {
        path: 'system/roles',
        name: 'SystemRoles',
        component: () => import('@/views/system/role/index.vue'),
        meta: { title: '角色管理', icon: 'Key', permission: 'system:roles' },
      },
      {
        path: 'base-data/items',
        name: 'BaseDataItems',
        component: () => import('@/views/baseData/items/index.vue'),
        meta: { title: '检测项目', icon: 'List', permission: 'base-data:items' },
      },
      {
        path: 'base-data/standards',
        name: 'BaseDataStandards',
        component: () => import('@/views/baseData/standards/index.vue'),
        meta: { title: '检测标准', icon: 'Document', permission: 'base-data:standards' },
      },
      {
        path: 'equipment/list',
        name: 'EquipmentList',
        component: () => import('@/views/baseData/equipment/index.vue'),
        meta: { title: '仪器设备', icon: 'Monitor', permission: 'base-data:equipment' },
      },
      {
        path: 'base-data/reagents',
        name: 'BaseDataReagents',
        component: () => import('@/views/baseData/reagents/index.vue'),
        meta: { title: '物资管理', icon: 'Box', permission: 'base-data:reagents' },
      },
      {
        path: 'form/templates',
        name: 'TemplateManage',
        component: () => import('@/views/baseData/templateManage/index.vue'),
        meta: { title: '表单模板管理', icon: 'Document', permission: 'base-data:items' },
      },
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
        meta: { title: '任务委托', icon: 'Edit', permission: 'business:task-order' },
      },
      {
        path: 'business/contract-review',
        name: 'ContractReview',
        component: () => import('@/views/business/contractReview/index.vue'),
        meta: { title: '合同评审', icon: 'Document', permission: 'business:contract-review' },
      },
      {
        path: 'business/qc-task',
        name: 'QCTask',
        component: () => import('@/views/business/qcTask/index.vue'),
        meta: { title: '质控任务', icon: 'DataAnalysis', permission: 'business:qc-task' },
      },
      {
        path: 'business/sampling-schedule',
        name: 'SamplingSchedule',
        component: () => import('@/views/business/samplingSchedule/index.vue'),
        meta: { title: '采样调度', icon: 'Calendar', permission: 'business:sampling-schedule' },
      },
      {
        path: 'business/field-sampling',
        name: 'FieldSampling',
        component: () => import('@/views/business/fieldSampling/index.vue'),
        meta: { title: '现场采样', icon: 'Location', permission: 'business:field-sampling' },
      },
      {
        path: 'business/sample-receiving',
        name: 'SampleReceiving',
        component: () => import('@/views/business/sampleReceiving/index.vue'),
        meta: { title: '样品接收', icon: 'Box', permission: 'business:sample-receiving' },
      },
      {
        path: 'business/task-assign',
        name: 'TaskAssign',
        component: () => import('@/views/business/taskAssign/index.vue'),
        meta: { title: '任务分配', icon: 'UserFilled', permission: 'business:task-assign' },
      },
      {
        path: 'business/data-entry',
        name: 'DataEntry',
        component: () => import('@/views/business/dataEntry/index.vue'),
        meta: { title: '数据录入', icon: 'EditPen', permission: 'business:data-entry' },
      },
      {
        path: 'business/lab-sheet-editor',
        name: 'LabSheetEditorPage',
        component: () => import('@/views/business/labSheetEditor/index.vue'),
        meta: { title: '检验单编辑', icon: 'Document', noSidebar: true },
      },
      {
        path: 'business/sampling-sheet-editor',
        name: 'SamplingSheetEditorPage',
        component: () => import('@/views/business/samplingSheetEditor/index.vue'),
        meta: { title: '采样单编辑', icon: 'Document', noSidebar: true },
      },
      {
        path: 'business/data-review',
        name: 'DataReview',
        component: () => import('@/views/business/dataReview/index.vue'),
        meta: { title: '数据复核', icon: 'Search', permission: 'business:data-review' },
      },
      {
        path: 'business/data-audit',
        name: 'DataAudit',
        component: () => import('@/views/business/dataAudit/index.vue'),
        meta: { title: '数据审核', icon: 'Finished', permission: 'business:data-audit' },
      },
      {
        path: 'business/report-prepare',
        name: 'ReportPrepare',
        component: () => import('@/views/business/reportPrepare/index.vue'),
        meta: { title: '报告编制', icon: 'Document', permission: 'business:report-prepare' },
      },
      {
        path: 'business/report-review',
        name: 'ReportReview',
        component: () => import('@/views/business/reportReview/index.vue'),
        meta: { title: '报告复核', icon: 'Search', permission: 'business:report-review' },
      },
      {
        path: 'business/report-audit',
        name: 'ReportAudit',
        component: () => import('@/views/business/reportAudit/index.vue'),
        meta: { title: '报告审核', icon: 'Finished', permission: 'business:report-audit' },
      },
      {
        path: 'business/report-sign',
        name: 'ReportSign',
        component: () => import('@/views/business/reportSign/index.vue'),
        meta: { title: '报告签发', icon: 'Stamp', permission: 'business:report-sign' },
      },
      {
        path: 'business/report-print',
        name: 'ReportPrint',
        component: () => import('@/views/business/reportPrint/index.vue'),
        meta: { title: '报告打印发放', icon: 'Printer', permission: 'business:report-print' },
      },
      {
        path: 'business/project-archive',
        name: 'ProjectArchive',
        component: () => import('@/views/business/projectArchive/index.vue'),
        meta: { title: '项目归档', icon: 'FolderOpened', permission: 'business:project-archive' },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { title: '无权访问', noAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '页面未找到', noAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

type PermissionMeta = string | string[] | undefined

function checkPermission(userStore: ReturnType<typeof useUserStore>, meta: PermissionMeta): { ok: boolean; required: string } {
  if (!meta) return { ok: true, required: '' }
  if (Array.isArray(meta)) {
    const ok = userStore.hasAnyPermission(meta)
    return { ok, required: meta.join(' 或 ') }
  }
  return { ok: userStore.hasPermission(meta), required: meta }
}

router.beforeEach((to, _from, next) => {
  if (to.meta.noAuth) {
    next()
    return
  }

  const userStore = useUserStore()

  if (!userStore.token) {
    next('/login')
    return
  }

  if (userStore.isExpired) {
    userStore.clearAuth()
    next('/login')
    return
  }

  const { ok, required } = checkPermission(userStore, to.meta.permission as PermissionMeta)
  if (!ok) {
    ElMessage.warning(`您没有访问此页面的权限${required ? '（需要：' + required + '）' : ''}`)
    next({ path: '/403', query: { from: to.fullPath, perm: required } })
    return
  }

  next()
})

export default router