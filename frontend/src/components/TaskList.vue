<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>{{ title }}</span>
        <el-tag v-if="tasks.length" type="info" size="small">{{ tasks.length }} 项</el-tag>
      </div>
    </template>

    <el-empty v-if="!tasks.length" description="暂无待办任务" />

    <el-table v-else :data="tasks" stripe v-loading="loading" @row-click="handleRowClick" style="cursor: pointer;">
      <el-table-column label="任务标题" min-width="180">
        <template #default="{ row }">
          <div class="task-title">{{ row.title }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="node_name" label="当前节点" width="120" />
      <el-table-column prop="create_time" label="到达时间" width="160" />
      <el-table-column label="紧急程度" width="90">
        <template #default="{ row }">
          <el-tag :type="urgencyType(row.urgency)" size="small">{{ row.urgency || '普通' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click.stop="handleAction(row)">处理</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { type PropType } from 'vue'

interface TaskItem {
  id: number
  title: string
  node_name: string
  create_time: string
  urgency?: string
  [key: string]: unknown
}

defineProps({
  title: { type: String, default: '待办任务' },
  tasks: { type: Array as PropType<TaskItem[]>, default: () => [] },
  loading: { type: Boolean, default: false },
})

const emit = defineEmits<{
  action: [task: TaskItem]
  'row-click': [task: TaskItem]
}>()

function urgencyType(urgency?: string): 'danger' | 'warning' | 'info' {
  if (urgency === '紧急') return 'danger'
  if (urgency === '重要') return 'warning'
  return 'info'
}

function handleRowClick(task: TaskItem) {
  emit('row-click', task)
}

function handleAction(task: TaskItem) {
  emit('action', task)
}
</script>

<style scoped>
.task-title {
  font-weight: 500;
}
</style>