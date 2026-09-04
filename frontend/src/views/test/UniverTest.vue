<template>
  <div style="padding: 20px; height: 100vh; background: #fff;">
    <h2>Univer 渲染测试 v4 (preset factory)</h2>
    <div id="univer-test-container" style="width: 100%; height: calc(100vh - 100px); border: 2px solid red;"></div>
    <pre id="debug-log" style="position:fixed;right:10px;top:10px;width:400px;height:300px;background:#000;color:#0f0;font-size:11px;overflow:auto;border:1px solid #fff;"></pre>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { Univer, LocaleType } from '@univerjs/core'
import { FUniver } from '@univerjs/core/facade'
import { UniverSheetsCorePreset } from '@univerjs/presets/preset-sheets-core'
import zhCN from '@univerjs/preset-sheets-core/locales/zh-CN'
import '@univerjs/design/lib/index.css'
import '@univerjs/ui/lib/index.css'

const logs: string[] = []
function log(msg: string) {
  logs.push(msg)
  console.log(msg)
  const el = document.getElementById('debug-log')
  if (el) el.textContent = logs.join('\n')
}

onMounted(() => {
  const el = document.getElementById('univer-test-container')
  if (!el) return

  window.addEventListener('error', (e) => log(`[window-error] ${e.message}`))
  window.addEventListener('unhandledrejection', (e) => log(`[unhandled-rejection] ${e.reason}`))

  try {
    const preset = UniverSheetsCorePreset({ container: el })
    log(`[ok] preset factory called, ${preset.plugins?.length || 0} plugins`)
    preset.plugins?.forEach((p: any, i: number) => {
      if (Array.isArray(p)) log(`  preset[${i}]: tuple len=${p.length} -> ${p[0]?.pluginName || p[0]?.name || '?'}`)
      else log(`  preset[${i}]: ${typeof p} -> ${p?.pluginName || p?.name || 'undefined'}`)
    })

    const univer = new Univer({
      locale: LocaleType.ZH_CN,
      locales: { [LocaleType.ZH_CN]: zhCN },
    })
    log('[ok] Univer instance created')

    preset.plugins.forEach((p: any, idx: number) => {
      try {
        if (Array.isArray(p)) {
          const [Cls, cfg] = p
          if (Cls) univer.registerPlugin(Cls, cfg)
          else log(`  [WARN] preset[${idx}] array has undefined class`)
        } else if (p) {
          univer.registerPlugin(p)
        } else {
          log(`  [WARN] preset[${idx}] is undefined`)
        }
      } catch (e: any) {
        log(`  [ERR] preset[${idx}]: ${e?.message}`)
      }
    })
    log('[ok] all preset plugins registered')

    const api = FUniver.newAPI(univer)
    log('[ok] FUniver API created')

    const sheetId = 'sheet-test'
    const wb = api.createWorkbook({
      id: 'wb-test',
      name: 'Test',
      sheetOrder: [sheetId],
      sheets: {
        [sheetId]: {
          id: sheetId,
          name: 'Sheet1',
          rowCount: 20,
          columnCount: 10,
          cellData: {},
          columnData: {},
          rowData: {},
        },
      },
    })
    log(`[ok] Workbook created: ${wb?.getId?.()}`)

    setTimeout(() => {
      log(`[after 800ms] children: ${el.children.length}`)
      for (let i = 0; i < el.children.length; i++) {
        const c = el.children[i] as HTMLElement
        const cvs = c.querySelectorAll('canvas')
        log(`  child[${i}]: ${c.tagName} w=${c.clientWidth} h=${c.clientHeight} canvases=${cvs.length}`)
        cvs.forEach((cv: HTMLCanvasElement, j) => {
          log(`    canvas[${j}]: ${cv.width}x${cv.height}`)
        })
      }
      log(`[total document canvases] ${document.querySelectorAll('canvas').length}`)
    }, 800)

  } catch (err: any) {
    log(`[FATAL] ${err?.message || err}`)
    log(err?.stack || '')
  }
})
</script>