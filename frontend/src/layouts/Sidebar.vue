<template>
  <div class="sidebar">
    <div class="sidebar-logo">
      <img v-if="!collapsed" src="/src/assets/logo.svg" alt="LIMS" class="logo-img" />
      <span v-else class="logo-mini">L</span>
    </div>
    <el-menu
      :default-active="activeMenu"
      :collapse="collapsed"
      :collapse-transition="false"
      background-color="#304156"
      text-color="#bfcbd9"
      active-text-color="#409eff"
      router
    >
      <el-menu-item index="/dashboard">
        <el-icon><Odometer /></el-icon>
        <template #title>工作台</template>
      </el-menu-item>

      <el-sub-menu index="system" v-if="checkAny(['system:users', 'system:depts', 'system:roles'])">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>系统管理</span>
        </template>
        <el-menu-item index="/system/users" v-if="checkPerm('system:users')">
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
        <el-menu-item index="/system/depts" v-if="checkPerm('system:depts')">
          <el-icon><OfficeBuilding /></el-icon>
          <template #title>部门管理</template>
        </el-menu-item>
        <el-menu-item index="/system/roles" v-if="checkPerm('system:roles')">
          <el-icon><Key /></el-icon>
          <template #title>角色管理</template>
        </el-menu-item>
      </el-sub-menu>

      <el-sub-menu index="base-data" v-if="checkAny(['base-data:items', 'base-data:standards', 'base-data:equipment', 'base-data:reagents'])">
        <template #title>
          <el-icon><FolderOpened /></el-icon>
          <span>基础数据</span>
        </template>
        <el-menu-item index="/base-data/items" v-if="checkPerm('base-data:items')">
          <el-icon><List /></el-icon>
          <template #title>检测项目</template>
        </el-menu-item>
        <el-menu-item index="/base-data/standards" v-if="checkPerm('base-data:standards')">
          <el-icon><Document /></el-icon>
          <template #title>检测标准</template>
        </el-menu-item>
        <el-menu-item index="/base-data/equipment" v-if="checkPerm('base-data:equipment')">
          <el-icon><Monitor /></el-icon>
          <template #title>仪器设备</template>
        </el-menu-item>
        <el-menu-item index="/base-data/reagents" v-if="checkPerm('base-data:reagents')">
          <el-icon><Box /></el-icon>
          <template #title>试剂耗材</template>
        </el-menu-item>
      </el-sub-menu>

      <el-sub-menu index="business" v-if="hasAnyBusinessPerm">
        <template #title>
          <el-icon><TrendCharts /></el-icon>
          <span>商务流程</span>
        </template>
        <el-menu-item index="/business/pending-tasks">
          <el-icon><Bell /></el-icon>
          <template #title>待办任务</template>
        </el-menu-item>
        <el-menu-item index="/business/task-order" v-if="checkPerm('business:task-order')">
          <el-icon><Edit /></el-icon>
          <template #title>任务委托</template>
        </el-menu-item>
        <el-menu-item index="/business/contract-review" v-if="checkPerm('business:contract-review')">
          <el-icon><Document /></el-icon>
          <template #title>合同评审</template>
        </el-menu-item>
        <el-menu-item index="/business/qc-task" v-if="checkPerm('business:qc-task')">
          <el-icon><DataAnalysis /></el-icon>
          <template #title>质控任务</template>
        </el-menu-item>
        <el-menu-item index="/business/sampling-schedule" v-if="checkPerm('business:sampling-schedule')">
          <el-icon><Calendar /></el-icon>
          <template #title>采样调度</template>
        </el-menu-item>
        <el-menu-item index="/business/field-sampling" v-if="checkPerm('business:field-sampling')">
          <el-icon><Location /></el-icon>
          <template #title>现场采样</template>
        </el-menu-item>
        <el-menu-item index="/business/sample-receiving" v-if="checkPerm('business:sample-receiving')">
          <el-icon><Box /></el-icon>
          <template #title>样品接收</template>
        </el-menu-item>
        <el-menu-item index="/business/task-assign" v-if="checkPerm('business:task-assign')">
          <el-icon><UserFilled /></el-icon>
          <template #title>任务分配</template>
        </el-menu-item>
        <el-menu-item index="/business/data-entry" v-if="checkPerm('business:data-entry')">
          <el-icon><EditPen /></el-icon>
          <template #title>数据录入</template>
        </el-menu-item>
        <el-menu-item index="/business/data-review" v-if="checkPerm('business:data-review')">
          <el-icon><Search /></el-icon>
          <template #title>数据复核</template>
        </el-menu-item>
        <el-menu-item index="/business/data-audit" v-if="checkPerm('business:data-audit')">
          <el-icon><Finished /></el-icon>
          <template #title>数据审核</template>
        </el-menu-item>
        <el-menu-item index="/business/report-prepare" v-if="checkPerm('business:report-prepare')">
          <el-icon><Document /></el-icon>
          <template #title>报告编制</template>
        </el-menu-item>
        <el-menu-item index="/business/report-review" v-if="checkPerm('business:report-review')">
          <el-icon><Search /></el-icon>
          <template #title>报告复核</template>
        </el-menu-item>
        <el-menu-item index="/business/report-audit" v-if="checkPerm('business:report-audit')">
          <el-icon><Finished /></el-icon>
          <template #title>报告审核</template>
        </el-menu-item>
        <el-menu-item index="/business/report-sign" v-if="checkPerm('business:report-sign')">
          <el-icon><Stamp /></el-icon>
          <template #title>报告签发</template>
        </el-menu-item>
        <el-menu-item index="/business/report-print" v-if="checkPerm('business:report-print')">
          <el-icon><Printer /></el-icon>
          <template #title>报告打印发放</template>
        </el-menu-item>
        <el-menu-item index="/business/project-archive" v-if="checkPerm('business:project-archive')">
          <el-icon><FolderOpened /></el-icon>
          <template #title>项目归档</template>
        </el-menu-item>
      </el-sub-menu>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import {
  Odometer, Setting, User, OfficeBuilding, Key, FolderOpened, List, Document,
  Monitor, Box, TrendCharts, Bell, Edit, DataAnalysis, Calendar, Location,
  UserFilled, EditPen, Search, Finished, Stamp, Printer,
} from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{ collapsed: boolean }>(), {
  collapsed: false,
})

const route = useRoute()
const userStore = useUserStore()
const activeMenu = computed(() => route.path)

function checkPerm(code: string): boolean {
  return userStore.hasPermission(code)
}

function checkAny(codes: string[]): boolean {
  if (userStore.isAdmin || userStore.permissions.includes('*')) return true
  return codes.some(c => userStore.hasPermission(c))
}

const businessPerms = [
  'business:task-order', 'business:contract-review', 'business:qc-task',
  'business:sampling-schedule', 'business:field-sampling', 'business:sample-receiving',
  'business:task-assign', 'business:data-entry', 'business:data-review', 'business:data-audit',
  'business:report-prepare', 'business:report-review', 'business:report-audit',
  'business:report-sign', 'business:report-print', 'business:project-archive',
]
const hasAnyBusinessPerm = computed(() => {
  if (userStore.isAdmin || userStore.permissions.includes('*')) return true
  // 待办任务对所有人可见
  return businessPerms.some(c => userStore.hasPermission(c))
})
</script>

<style scoped>
.sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.sidebar-logo {
  height: 50px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}
.logo-img {
  height: 32px;
}
.logo-mini {
  color: #fff;
  font-size: 24px;
  font-weight: bold;
}
.el-menu {
  border-right: none;
  flex: 1;
  overflow-y: auto;
}
</style>