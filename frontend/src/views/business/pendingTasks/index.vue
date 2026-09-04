<template>
  <div class="pending-tasks">
    <TaskList
      :title="isAdminUser ? '部门待办（管理员·全部部门）' : '部门待办'"
      :tasks="deptTasks"
      :loading="deptLoading"
      @action="openApproval"
      @row-click="showDetail"
    />

    <TaskList
      title="我的待办"
      class="my-tasks"
      :tasks="myTasks"
      :loading="myLoading"
      @action="openApproval"
      @row-click="showDetail"
    />

    <el-card class="timeline-card" v-if="selectedInstanceId">
      <template #header><span>流程进度</span></template>
      <ProcessTimeline :nodes="timelineNodes" />
    </el-card>
    <el-empty v-else class="timeline-card" description="请选择待办任务查看流程" />

    <ApprovalDialog ref="approvalDialogRef" :reject-targets="nodeDefs" @submit="handleApprovalSubmit" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import TaskList from '@/components/TaskList.vue'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import ProcessTimeline from '@/components/ProcessTimeline.vue'
import {
  getPendingTasksByDept,
  getPendingTasksByUser,
  getProcessHistory,
  getNodeDefinitions,
  approveTask,
  rejectTask,
} from '@/api/workflow'
import type { PendingTask, TimelineNode, NodeDefinition } from '@/api/workflow'
import { isAdmin } from '@/utils/permission'

const deptLoading = ref(false)
const myLoading = ref(false)
const deptTasks = ref<PendingTask[]>([])
const myTasks = ref<PendingTask[]>([])
const selectedInstanceId = ref<number | null>(null)
const timelineNodes = ref<TimelineNode[]>([])
const nodeDefs = ref<NodeDefinition[]>([])
const isAdminUser = isAdmin()
const approvalDialogRef = ref<InstanceType<typeof ApprovalDialog>>()
let currentTaskId = 0

onMounted(async () => {
  await loadTasks()
  try {
    const res = await getNodeDefinitions()
    nodeDefs.value = res.data
  } catch {
    // nodes optional
  }
})

async function loadTasks() {
  deptLoading.value = true
  myLoading.value = true
  try {
    const deptRes = await getPendingTasksByDept()
    deptTasks.value = deptRes.data
  } catch { deptTasks.value = [] }
  finally { deptLoading.value = false }

  try {
    const myRes = await getPendingTasksByUser()
    myTasks.value = myRes.data
  } catch { myTasks.value = [] }
  finally { myLoading.value = false }
}

function openApproval(task: unknown) {
  const t = task as PendingTask
  currentTaskId = t.id
  approvalDialogRef.value?.open()
}

async function showDetail(task: unknown) {
  const t = task as PendingTask
  if (!t.process_instance_id) return
  selectedInstanceId.value = t.process_instance_id
  try {
    const res = await getProcessHistory(t.process_instance_id)
    timelineNodes.value = res.data as TimelineNode[]
  } catch {
    timelineNodes.value = []
  }
}

async function handleApprovalSubmit(data: { action: string; comment: string; reject_target?: string }) {
  if (data.action === 'approve') {
    await approveTask(currentTaskId, { comment: data.comment })
    ElMessage.success('审批通过')
  } else {
    await rejectTask(currentTaskId, { comment: data.comment })
    ElMessage.success('已驳回')
  }
  await loadTasks()
}
</script>

<style scoped>
.pending-tasks {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.timeline-card {
  min-height: 120px;
}
</style>