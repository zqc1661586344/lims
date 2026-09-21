<template>
  <div v-if="progress">
    <div class="task-process-bar" :class="{ sticky: sticky }">
      <div class="process-header">
        <span class="process-title">📋 流程进度</span>
        <span class="process-status" :class="'status-' + progress.status">
          {{ statusText }}
        </span>
        <span class="process-percent" v-if="progress.total_nodes > 0">
          {{ completedCount }}/{{ progress.total_nodes }} 步
        </span>
      </div>
      <el-steps
        :active="progress.current_node_index >= 0 ? progress.current_node_index + 1 : 0"
        process-status="process"
        finish-status="success"
        align-center
        simple
        :space-auto="false"
      >
        <el-step
          v-for="node in visibleNodes"
          :key="node.code"
          :title="node.name"
          :status="stepStatus(node)"
          :icon="stepIcon(node)"
        >
          <template #icon>
            <div class="step-icon" :class="'step-' + node.status">
              <el-icon v-if="node.status === 'completed'"><Check /></el-icon>
              <el-icon v-else-if="node.status === 'rejected'"><Close /></el-icon>
              <el-icon v-else-if="node.status === 'current'" class="current-pulse"><Loading /></el-icon>
              <span v-else>{{ node.index + 1 }}</span>
            </div>
          </template>
        </el-step>
      </el-steps>

      <div class="compact-view" v-if="!expanded">
        <el-button link size="small" @click="expanded = true">展开详情</el-button>
      </div>
      <div class="detail-view" v-else>
        <el-divider style="margin: 12px 0" />
        <div class="detail-nodes">
          <div
            v-for="node in progress.nodes"
            :key="node.code"
            class="detail-node"
            :class="'detail-' + node.status"
          >
            <div class="detail-left">
              <div class="detail-dot" :class="'dot-' + node.status"></div>
              <span class="detail-name">{{ node.name }}</span>
            </div>
            <div class="detail-right">
              <el-tag v-if="node.status === 'completed'" type="success" size="small">已完成</el-tag>
              <el-tag v-else-if="node.status === 'rejected'" type="danger" size="small">已驳回</el-tag>
              <el-tag v-else-if="node.status === 'current'" type="primary" size="small">进行中</el-tag>
              <el-tag v-else-if="node.status === 'skipped'" type="info" size="small">已跳过</el-tag>
              <el-tag v-else type="info" size="small" effect="plain">待处理</el-tag>
              <span v-if="node.comment" class="detail-comment">
                <el-icon><ChatDotRound /></el-icon>
              </span>
            </div>
          </div>
        </div>
        <el-button link size="small" @click="expanded = false">收起</el-button>
      </div>
    </div>
  </div>

  <div class="task-process-bar empty" v-if="resolvedBusinessId && !progress && !loading">
    <el-alert type="info" :closable="false" show-icon title="该任务尚未启动流程" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { getWorkflowProgress, type WorkflowProgress, type ProgressNode } from '@/api/workflow'
import { Check, Close, Loading, ChatDotRound } from '@element-plus/icons-vue'

const props = defineProps<{
  businessType?: string
  businessId?: number | null
  items?: Array<{ task_order_id: number }>
  sticky?: boolean
  selectedId?: number | null
}>()

const resolvedBusinessId = computed(() => {
  if (props.selectedId != null) return props.selectedId
  if (props.businessId != null) return props.businessId
  if (props.items && props.items.length > 0) {
    const id = props.items[0]?.task_order_id
    return id || null
  }
  return null
})

const progress = ref<WorkflowProgress | null>(null)
const loading = ref(false)
const expanded = ref(false)

const completedCount = computed(() => {
  if (!progress.value) return 0
  return progress.value.nodes.filter(n => n.status === 'completed' || n.status === 'skipped').length
})

const statusText = computed(() => {
  if (!progress.value) return ''
  const s = progress.value.status
  if (s === 'completed') return '✅ 已完成'
  if (s === 'terminated') return '⛔ 已终止'
  return '🔄 进行中'
})

const visibleNodes = computed(() => {
  if (!progress.value) return []
  return progress.value.nodes
})

function stepStatus(node: ProgressNode): 'success' | 'error' | 'process' | 'wait' | 'finish' {
  if (node.status === 'completed' || node.status === 'skipped') return 'success'
  if (node.status === 'rejected') return 'error'
  if (node.status === 'current') return 'process'
  return 'wait'
}

function stepIcon(node: ProgressNode) {
  return ''
}

async function loadProgress() {
  const bt = props.businessType || 'task_order'
  const id = resolvedBusinessId.value
  
  if (!id) {
    progress.value = null
    return
  }
  loading.value = true
  try {
    const res = await getWorkflowProgress(bt, id)
    progress.value = res.data || null
  } catch {
    progress.value = null
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.businessType, resolvedBusinessId.value],
  () => loadProgress(),
  { immediate: false }
)

onMounted(() => loadProgress())

defineExpose({ refresh: loadProgress })
</script>

<style scoped>
.task-process-bar {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 16px;
  z-index: 10;
}

.task-process-bar.sticky {
  position: sticky;
  top: 0;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
}

.task-process-bar.empty {
  padding: 8px 12px;
}

.process-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.process-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.process-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}

.process-status.status-completed {
  background: var(--el-color-success-light-9);
  color: var(--el-color-success);
}

.process-status.status-terminated {
  background: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

.process-percent {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-left: auto;
}

.step-icon {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  border: 2px solid var(--el-border-color);
  background: var(--el-bg-color);
  color: var(--el-text-color-secondary);
}

.step-icon.step-completed {
  background: var(--el-color-success);
  border-color: var(--el-color-success);
  color: #fff;
}

.step-icon.step-current {
  background: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: #fff;
}

.step-icon.step-rejected {
  background: var(--el-color-danger);
  border-color: var(--el-color-danger);
  color: #fff;
}

.current-pulse {
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.compact-view {
  margin-top: 8px;
}

.detail-nodes {
  max-height: 320px;
  overflow-y: auto;
}

.detail-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 8px;
  border-radius: 4px;
  margin-bottom: 4px;
  font-size: 13px;
}

.detail-node.detail-current {
  background: var(--el-color-primary-light-9);
}

.detail-node.detail-completed {
  background: var(--el-color-success-light-9);
}

.detail-node.detail-rejected {
  background: var(--el-color-danger-light-9);
}

.detail-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--el-text-color-placeholder);
}

.detail-dot.dot-completed {
  background: var(--el-color-success);
}

.detail-dot.dot-current {
  background: var(--el-color-primary);
}

.detail-dot.dot-rejected {
  background: var(--el-color-danger);
}

.detail-dot.dot-skipped {
  background: var(--el-text-color-secondary);
}

.detail-comment {
  margin-left: 6px;
  color: var(--el-color-warning);
}
</style>