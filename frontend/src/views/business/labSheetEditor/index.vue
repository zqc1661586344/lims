<script setup lang="ts">
import { computed, onMounted, ref, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Loading } from '@element-plus/icons-vue'
import { Univer, LocaleType } from '@univerjs/core'
import { FUniver } from '@univerjs/core/facade'
import { UniverSheetsCorePreset } from '@univerjs/presets/preset-sheets-core'
import zhCN from '@univerjs/preset-sheets-core/locales/zh-CN'
import {
  listLabSheets,
  createLabSheet,
  updateLabSheet,
  type LabSheet,
} from '@/api/business/labSheet'

const route = useRoute()
const router = useRouter()

const taskOrderId = computed(() => Number(route.query.task_order_id) || 0)
const testItemId = computed(() => Number(route.query.test_item_id) || 0)
const nodeCode = computed(() => (route.query.node_code as string) || 'node_data_entry')

const sheetId = ref<number | null>(null)
const saving = ref(false)
const containerRef = ref<HTMLElement>()
const univerRef = shallowRef<Univer | null>(null)
const apiRef = shallowRef<FUniver | null>(null)
let wbRef: any = null
let initialized = false

const title = computed(() => `检验单 - 委托 #${taskOrderId.value}`)

async function loadAndInit() {
  if (!taskOrderId.value || !testItemId.value) {
    ElMessage.error('缺少参数：委托ID或项目ID')
    return
  }

  let existingData: any = null

  try {
    const res = await listLabSheets({
      task_order_id: String(taskOrderId.value),
      node_code: nodeCode.value,
    })
    const list: LabSheet[] = res.data || []
    const existing = list.find((s) => s.test_item_id === testItemId.value)
    if (existing) {
      sheetId.value = existing.id
      existingData = existing.sheet_data
      console.log('[labSheetEditor] loaded existing sheet:', sheetId.value, 'data keys:', Object.keys(existingData || {}))
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载检验单失败')
    return
  }

  await initUniver(existingData)
}

async function initUniver(existingData: any) {
  if (initialized) return
  if (!containerRef.value) {
    await new Promise(r => setTimeout(r, 200))
  }
  const el = containerRef.value!

  for (let i = 0; i < 20 && (el.clientHeight < 30 || el.clientWidth < 30); i++) {
    await new Promise(r => setTimeout(r, 100))
  }

  if (el.clientHeight < 30) {
    console.error('[labSheetEditor] container too small:', el.clientWidth, 'x', el.clientHeight)
    return
  }

  initialized = true

  try {
    const preset = UniverSheetsCorePreset({ container: el })

    const univer = new Univer({
      locale: LocaleType.ZH_CN,
      locales: { [LocaleType.ZH_CN]: zhCN },
    })

    preset.plugins.forEach((p: any) => {
      if (Array.isArray(p)) {
        const [Cls, cfg] = p
        if (Cls) univer.registerPlugin(Cls, cfg)
      } else if (p) {
        univer.registerPlugin(p)
      }
    })

    univerRef.value = univer
    const api = FUniver.newAPI(univer)
    apiRef.value = api

    let wbData: any
    if (existingData && typeof existingData === 'object' && Object.keys(existingData).length > 0 && existingData.sheets) {
      wbData = existingData
    } else {
      const sheetIdStr = `sheet-${Date.now()}`
      wbData = {
        id: `wb-${Date.now()}`,
        name: 'Sheet',
        sheetOrder: [sheetIdStr],
        sheets: {
          [sheetIdStr]: {
            id: sheetIdStr,
            name: 'Sheet1',
            rowCount: 50,
            columnCount: 15,
            cellData: {},
            columnData: {},
            rowData: {},
          },
        },
        locale: LocaleType.ZH_CN,
        creator: 'LIMS',
      }
    }

    wbRef = api.createWorkbook(wbData)
    console.log('[labSheetEditor] Univer initialized:', wbRef?.getId?.(), 'container:', el.clientWidth, 'x', el.clientHeight)
  } catch (e) {
    console.error('[labSheetEditor] init FAILED:', e)
    initialized = false
  }
}

async function handleSave() {
  if (!wbRef) { ElMessage.warning('表格未就绪'); return }
  const snapshot = (() => { try { return wbRef.save?.() || wbRef.getSnapshot?.() } catch { return null } })()
  if (!snapshot) { ElMessage.warning('无数据可保存'); return }

  saving.value = true
  try {
    if (sheetId.value) {
      await updateLabSheet(sheetId.value, { sheet_data: snapshot, node_code: nodeCode.value })
    } else {
      const res = await createLabSheet({
        task_order_id: taskOrderId.value,
        test_item_id: testItemId.value,
        sheet_data: snapshot,
        node_code: nodeCode.value,
      })
      sheetId.value = res.data?.id ?? null
    }
    ElMessage.success('保存成功')
    setTimeout(() => { router.push({ name: 'DataEntry' }) }, 800)
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function handleBack() { router.go(-1) }

onMounted(async () => {
  console.log('[labSheetEditor] onMounted, containerRef=', containerRef.value, 'size=', containerRef.value?.clientWidth, 'x', containerRef.value?.clientHeight)
  await loadAndInit()
})
</script>

<template>
  <div class="sheet-wrap">
    <div class="sheet-topbar">
      <div class="topbar-left">
        <el-button :icon="ArrowLeft" text @click="handleBack">返回</el-button>
        <el-divider direction="vertical" />
        <span class="topbar-title">{{ title }}</span>
        <el-tag class="topbar-tag" type="info">项目 #{{ testItemId }}</el-tag>
        <el-tag v-if="sheetId" class="topbar-tag" type="success">检验单 #{{ sheetId }}</el-tag>
        <el-tag v-else class="topbar-tag" type="warning">新建</el-tag>
      </div>
      <div class="topbar-right">
        <el-button type="primary" :loading="saving" @click="handleSave">保存检验单</el-button>
      </div>
    </div>

    <div ref="containerRef" class="sheet-container"></div>
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
.sheet-container {
  width: 100%;
  height: calc(100vh - 32px - 60px);
}
</style>