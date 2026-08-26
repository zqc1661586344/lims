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

      <!-- System Management -->
      <el-sub-menu index="system">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>系统管理</span>
        </template>
        <el-menu-item index="/system/users">
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
        <el-menu-item index="/system/depts">
          <el-icon><Organization /></el-icon>
          <template #title>部门管理</template>
        </el-menu-item>
        <el-menu-item index="/system/roles">
          <el-icon><Key /></el-icon>
          <template #title>角色管理</template>
        </el-menu-item>
      </el-sub-menu>

      <!-- Base Data -->
      <el-sub-menu index="base-data">
        <template #title>
          <el-icon><FolderOpened /></el-icon>
          <span>基础数据</span>
        </template>
        <el-menu-item index="/base-data/items">
          <el-icon><List /></el-icon>
          <template #title>检测项目</template>
        </el-menu-item>
        <el-menu-item index="/base-data/standards">
          <el-icon><Document /></el-icon>
          <template #title>检测标准</template>
        </el-menu-item>
        <el-menu-item index="/base-data/equipment">
          <el-icon><Monitor /></el-icon>
          <template #title>仪器设备</template>
        </el-menu-item>
        <el-menu-item index="/base-data/reagents">
          <el-icon><Box /></el-icon>
          <template #title>试剂耗材</template>
        </el-menu-item>
      </el-sub-menu>

      <!-- Business Process -->
      <el-sub-menu index="business">
        <template #title>
          <el-icon><TrendCharts /></el-icon>
          <span>商务流程</span>
        </template>
        <el-menu-item index="/business/pending-tasks">
          <el-icon><Bell /></el-icon>
          <template #title>待办任务</template>
        </el-menu-item>
        <el-menu-item index="/business/task-order">
          <el-icon><Edit /></el-icon>
          <template #title>任务委托</template>
        </el-menu-item>
        <el-menu-item index="/business/contract-review">
          <el-icon><Document /></el-icon>
          <template #title>合同评审</template>
        </el-menu-item>
        <el-menu-item index="/business/qc-task">
          <el-icon><DataAnalysis /></el-icon>
          <template #title>质控任务</template>
        </el-menu-item>
        <el-menu-item index="/business/sampling-schedule">
          <el-icon><Calendar /></el-icon>
          <template #title>采样调度</template>
        </el-menu-item>
        <el-menu-item index="/business/field-sampling">
          <el-icon><Location /></el-icon>
          <template #title>现场采样</template>
        </el-menu-item>
        <el-menu-item index="/business/sample-receiving">
          <el-icon><Box /></el-icon>
          <template #title>样品接收</template>
        </el-menu-item>
        <el-menu-item index="/business/task-assign">
          <el-icon><UserFilled /></el-icon>
          <template #title>任务分配</template>
        </el-menu-item>
        <el-menu-item index="/business/data-entry">
          <el-icon><EditPen /></el-icon>
          <template #title>数据录入</template>
        </el-menu-item>
        <el-menu-item index="/business/data-review">
          <el-icon><Search /></el-icon>
          <template #title>数据复核</template>
        </el-menu-item>
        <el-menu-item index="/business/data-audit">
          <el-icon><Finished /></el-icon>
          <template #title>数据审核</template>
        </el-menu-item>
        <el-menu-item index="/business/report-prepare">
          <el-icon><Document /></el-icon>
          <template #title>报告编制</template>
        </el-menu-item>
        <el-menu-item index="/business/report-review">
          <el-icon><Search /></el-icon>
          <template #title>报告复核</template>
        </el-menu-item>
        <el-menu-item index="/business/report-audit">
          <el-icon><Finished /></el-icon>
          <template #title>报告审核</template>
        </el-menu-item>
      </el-sub-menu>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Odometer, Setting, User, Organization, Key, FolderOpened, List, Document, Monitor, Box, TrendCharts, Bell, Edit, DataAnalysis, Calendar, Location, UserFilled, EditPen, Search, Finished } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{ collapsed: boolean }>(), {
  collapsed: false,
})

const route = useRoute()
const activeMenu = computed(() => route.path)
</script>

<style scoped>
.sidebar {
  height: 100%;
}

.sidebar-logo {
  height: 50px;
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
}
</style>