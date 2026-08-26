<template>
  <div class="process-timeline">
    <el-timeline>
      <el-timeline-item
        v-for="(node, index) in nodes"
        :key="node.code"
        :timestamp="node.time || ''"
        :type="nodeStatus(node, index)"
        :hollow="!node.active"
        placement="top"
        size="large"
      >
        <div class="timeline-node">
          <div class="node-header">
            <span class="node-name">{{ node.name }}</span>
            <el-tag v-if="node.status === 'completed'" type="success" size="small">已完成</el-tag>
            <el-tag v-else-if="node.status === 'rejected'" type="danger" size="small">已驳回</el-tag>
            <el-tag v-else-if="node.active" type="primary" size="small">进行中</el-tag>
            <el-tag v-else type="info" size="small">待处理</el-tag>
          </div>
          <div v-if="node.operator" class="node-operator">
            {{ node.operator }}
            <span v-if="node.dept"> · {{ node.dept }}</span>
          </div>
          <div v-if="node.comment" class="node-comment">
            <el-text type="warning" size="small">{{ node.comment }}</el-text>
          </div>
        </div>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<script setup lang="ts">
import { type PropType } from 'vue'

interface TimelineNode {
  code: string
  name: string
  time?: string
  status?: 'pending' | 'completed' | 'rejected'
  active?: boolean
  operator?: string
  dept?: string
  comment?: string
}

defineProps({
  nodes: { type: Array as PropType<TimelineNode[]>, default: () => [] },
})

function nodeStatus(node: TimelineNode, index: number): 'primary' | 'success' | 'danger' | 'info' {
  if (node.status === 'completed') return 'success'
  if (node.status === 'rejected') return 'danger'
  if (node.active) return 'primary'
  return 'info'
}
</script>

<style scoped>
.process-timeline {
  padding: 8px 0;
}
.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
}
.node-name {
  font-weight: 600;
  font-size: 14px;
}
.node-operator {
  color: #909399;
  font-size: 12px;
  margin-top: 2px;
}
.node-comment {
  margin-top: 4px;
  padding: 4px 8px;
  background: #fdf6ec;
  border-radius: 4px;
}
</style>