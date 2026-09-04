<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { Univer, LocaleType } from '@univerjs/core'
import { FUniver } from '@univerjs/core/facade'
import { UniverSheetsCorePreset } from '@univerjs/presets/preset-sheets-core'
import zhCN from '@univerjs/preset-sheets-core/locales/zh-CN'

const props = withDefaults(defineProps<{
  mode?: 'edit' | 'readonly'
  sheetData?: any
  height?: string
  title?: string
}>(), {
  mode: 'edit',
  height: '500px',
  title: '',
})

const emit = defineEmits<{
  (e: 'ready', api: FUniver): void
  (e: 'update:sheetData', snapshot: any): void
}>()

const containerRef = ref<HTMLElement>()
const univerRef = shallowRef<Univer | null>(null)
const apiRef = shallowRef<FUniver | null>(null)
let fWorkbook: any = null
let initialized = false

function buildDefaultWorkbookData(existingData?: any) {
  if (existingData && typeof existingData === 'object' && Object.keys(existingData).length > 0) {
    if (existingData.sheets && Object.keys(existingData.sheets).length > 0) {
      return existingData
    }
  }
  const sheetId = `sheet-${Date.now()}`
  return {
    id: `workbook-${Date.now()}`,
    name: props.title || 'Sheet',
    sheetOrder: [sheetId],
    sheets: {
      [sheetId]: {
        id: sheetId,
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
    createTime: new Date().toISOString(),
  }
}

async function init() {
  if (initialized) return
  if (!containerRef.value) {
    console.warn('[LabSheetEditor] init: containerRef null, deferring')
    await nextTick()
    if (!containerRef.value) return
  }

  const el = containerRef.value!

  for (let i = 0; i < 20 && (el.clientHeight < 30 || el.clientWidth < 30); i++) {
    await new Promise(r => setTimeout(r, 100))
  }

  if (el.clientHeight < 30 || el.clientWidth < 30) {
    console.error('[LabSheetEditor] container too small:', el.clientWidth, 'x', el.clientHeight)
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

    const wbData = buildDefaultWorkbookData(props.sheetData)
    fWorkbook = api.createWorkbook(wbData)

    if (props.mode === 'readonly') {
      try { fWorkbook?.setEditable?.(false) } catch {}
    }

    emit('ready', api)
    console.log('[LabSheetEditor] initialized OK:', fWorkbook?.getId?.(), 'mode:', props.mode)
  } catch (e) {
    console.error('[LabSheetEditor] init FAILED:', e)
    initialized = false
  }
}

function destroy() {
  initialized = false
  if (fWorkbook && apiRef.value) {
    try { apiRef.value.disposeUnit(fWorkbook.getId()) } catch {}
    fWorkbook = null
  }
  if (univerRef.value) {
    try { univerRef.value.dispose() } catch {}
    univerRef.value = null
  }
  apiRef.value = null
}

onBeforeUnmount(() => { destroy() })
onMounted(() => { init() })

watch(() => props.mode, (mode) => {
  if (fWorkbook) {
    try { fWorkbook.setEditable?.(mode !== 'readonly') } catch {}
  }
})

function getSnapshot() {
  try { return fWorkbook?.save?.() || fWorkbook?.getSnapshot?.() } catch { return null }
}

defineExpose({ init, destroy, getSnapshot })
</script>

<template>
  <div class="lab-sheet-editor">
    <div v-if="title" class="lab-sheet-title">{{ title }}</div>
    <div ref="containerRef" class="lab-sheet-container" :style="{ height }"></div>
  </div>
</template>

<style scoped>
.lab-sheet-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
}
.lab-sheet-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  padding: 8px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
  border-radius: 4px 4px 0 0;
  flex-shrink: 0;
}
.lab-sheet-container {
  width: 100%;
  position: relative;
  min-height: 300px;
  border: 1px solid #ebeef5;
  border-top: none;
  border-radius: 0 0 4px 4px;
  overflow: hidden;
}
.lab-sheet-container :deep(> *) {
  width: 100% !important;
  height: 100% !important;
}
</style>

<style>
.lab-sheet-container canvas {
  display: block !important;
}
</style>