<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import LabSheetEditor from '@/components/LabSheetEditor.vue'
import {
  listSamplingSheets,
  createSamplingSheet,
  updateSamplingSheet,
  type SamplingSheet,
} from '@/api/business/samplingSheet'

const route = useRoute()
const router = useRouter()

const taskOrderId = computed(() => Number(route.query.task_order_id) || 0)
const samplingPoint = computed(() => (route.query.sampling_point as string) || '')
const nodeCode = computed(() => (route.query.node_code as string) || 'node_field_sampling')

const mode = computed<'edit' | 'readonly'>(() => (route.query.mode as string) === 'readonly' ? 'readonly' : 'edit')

const sheetId = ref<number | null>(null)
const sheetData = ref<any>(null)
const saving = ref(false)
const editorRef = ref<InstanceType<typeof LabSheetEditor>>()

const title = computed(() => `采样单 - 委托 #${taskOrderId.value}`)

async function loadSheet() {
  if (!taskOrderId.value) {
    ElMessage.error('缺少参数：委托ID')
    return
  }
  try {
    const res = await listSamplingSheets({
      task_order_id: String(taskOrderId.value),
      node_code: nodeCode.value,
    })
    const list: SamplingSheet[] = (res.data?.items ?? res.data) || []
    let existing: SamplingSheet | undefined
    if (samplingPoint.value) {
      existing = list.find((s) => s.sampling_point === samplingPoint.value)
    } else {
      existing = list[0]
    }
    if (existing) {
      sheetId.value = existing.id
      sheetData.value = existing.sheet_data
    } else {
      sheetData.value = null
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载采样单失败')
  }
}

async function handleSave() {
  const snapshot = editorRef.value?.getSnapshot()
  if (!snapshot) { ElMessage.warning('无数据可保存'); return }
  saving.value = true
  try {
    if (sheetId.value) {
      await updateSamplingSheet(sheetId.value, { sheet_data: snapshot, node_code: nodeCode.value })
    } else {
      const res = await createSamplingSheet({
        task_order_id: taskOrderId.value,
        sampling_point: samplingPoint.value,
        sheet_data: snapshot,
        node_code: nodeCode.value,
      })
      sheetId.value = res.data?.id ?? null
    }
    ElMessage.success('保存成功')
    setTimeout(() => { router.push({ name: 'FieldSampling' }) }, 800)
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function handleBack() { router.go(-1) }

onMounted(async () => { await loadSheet() })
</script>

<template>
  <div class="sheet-wrap">
    <div class="sheet-topbar">
      <div class="topbar-left">
        <el-button :icon="ArrowLeft" text @click="handleBack">返回</el-button>
        <el-divider direction="vertical" />
        <span class="topbar-title">{{ title }}</span>
        <el-tag v-if="samplingPoint" class="topbar-tag" type="info">采样点: {{ samplingPoint }}</el-tag>
        <el-tag v-if="sheetId" class="topbar-tag" type="success">采样单 #{{ sheetId }}</el-tag>
        <el-tag v-else class="topbar-tag" type="warning">新建</el-tag>
        <el-tag v-if="mode === 'readonly'" class="topbar-tag" type="danger">只读</el-tag>
      </div>
      <div class="topbar-right">
        <el-button v-if="mode !== 'readonly'" type="primary" :loading="saving" @click="handleSave">保存采样单</el-button>
      </div>
    </div>

    <LabSheetEditor
      v-if="sheetData !== undefined"
      ref="editorRef"
      :mode="mode"
      :sheet-data="sheetData"
      :height="'calc(100vh - 32px - 60px)'"
    />
  </div>
</template>

<style scoped>
.sheet-wrap {
  width: 100%;
  height: calc(100vh - 32px);
  background: #fff;
  padding: 12px 16px;
  box-sizing: border-box;
}
.sheet-topbar {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.topbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.topbar-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-right: 8px;
}
.topbar-tag {
  margin-left: 4px;
}
</style>