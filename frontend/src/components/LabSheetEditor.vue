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
}>(), {
  mode: 'edit',
  height: '500px',
})

const emit = defineEmits<{
  (e: 'ready', api: FUniver): void
}>()

const containerRef = ref<HTMLElement>()
const univerRef = shallowRef<Univer | null>(null)
const apiRef = shallowRef<FUniver | null>(null)
let fWorkbook: any = null
let initialized = false

function buildWorkbookData(existingData?: any) {
  if (existingData && typeof existingData === 'object' && Object.keys(existingData).length > 0 && existingData.sheets) {
    return existingData
  }
  const sheetId = `sheet-${Date.now()}`
  return {
    id: `workbook-${Date.now()}`,
    name: 'Sheet',
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
  }
}

async function init() {
  if (initialized) return
  if (!containerRef.value) {
    await nextTick()
    if (!containerRef.value) return
  }

  const el = containerRef.value!
  for (let i = 0; i < 20 && (el.clientHeight < 30 || el.clientWidth < 30); i++) {
    await new Promise(r => setTimeout(r, 100))
  }
  if (el.clientHeight < 30) {
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

    fWorkbook = api.createWorkbook(buildWorkbookData(props.sheetData))
    if (props.mode === 'readonly') {
      try { fWorkbook?.setEditable?.(false) } catch {}
    }

    emit('ready', api)
    console.log('[LabSheetEditor] initialized, wb=', fWorkbook?.getId?.(), 'mode=', props.mode)
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

function getSnapshot() {
  try { return fWorkbook?.save?.() || fWorkbook?.getSnapshot?.() } catch { return null }
}

onMounted(() => { init() })
onBeforeUnmount(() => { destroy() })

watch(() => props.mode, (mode) => {
  if (fWorkbook) {
    try { fWorkbook.setEditable?.(mode !== 'readonly') } catch {}
  }
})

watch(() => props.sheetData, () => {
  if (initialized && fWorkbook && apiRef.value) {
    try {
      apiRef.value.disposeUnit(fWorkbook.getId())
    } catch {}
    fWorkbook = apiRef.value.createWorkbook(buildWorkbookData(props.sheetData))
    if (props.mode === 'readonly') {
      try { fWorkbook?.setEditable?.(false) } catch {}
    }
  }
})

defineExpose({ init, destroy, getSnapshot })
</script>

<template>
  <div ref="containerRef" class="lab-sheet-container" :style="{ height }"></div>
</template>

<style scoped>
.lab-sheet-container {
  width: 100%;
  position: relative;
  min-height: 300px;
}
</style>