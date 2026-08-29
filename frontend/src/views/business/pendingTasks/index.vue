<template>
  <div class="pending-tasks">
    <el-row :gutter="16">
      <el-col :span="8">
        <TaskList
          :title="isAdminUser ? '部门待办（管理员·全部部门）' : '部门待办'"
          :tasks="deptTasks"
          :loading="deptLoading"
          @action="openApproval"
          @row-click="showDetail"
        />
      </el-col>
      <el-col :span="8">
        <TaskList
          title="我的待办"
          :tasks="myTasks"
          :loading="myLoading"
          @action="openApproval"
          @row-click="showDetail"
        />
      </el-col>
      <el-col :span="8">
        <el-card v-if="selectedInstanceId">
          <template #header><span>流程进度</span></template>
          <ProcessTimeline :nodes="timelineNodes" />
        </el-card>
        <el-empty v-else description="请选择待办任务查看流程" />
      </el-col>
    </el-row>

    <ApprovalDialog ref="approvalDialogRef" :reject-targets="nodeDefs" @submit="handleApprovalSubmit" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import TaskList from '@/components/TaskList.vue'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import ProcessTimeline from '@/components/ProcessTimeline.vue'
import { getPendingTasksByDept, getPendingTasksByUser, getProcessHistory, getNodeDefinitions } from '@/api/business'
import type { PendingTask, TimelineNode, NodeDefinition } from '@/api/business'
import { approveTask, rejectTask } from '@/api/workflow'
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

function openApproval(task: PendingTask) {
  currentTaskId = task.id
  approvalDialogRef.value?.open()
}

async function showDetail(task: PendingTask) {
  if (!task.process_instance_id) return
  selectedInstanceId.value = task.process_instance_id
  try {
    const res = await getProcessHistory(task.process_instance_id)
    timelineNodes.value = res.data as TimelineNode[]
  } catch {
    timelineNodes.value = []
  }
}

async function handleApprovalSubmit(data: { action: string; comment: string; reject_target?: string }) {
  try {
    if (data.action === 'approve') {
      await approveTask(currentTaskId, { comment: data.comment })
      ElMessage.success('审批通过')
    } else {
      await rejectTask(currentTaskId, { comment: data.comment })
      ElMessage.success('已驳回')
    }
    await loadTasks()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  }
}
</script>