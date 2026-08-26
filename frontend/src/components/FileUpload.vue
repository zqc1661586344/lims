<template>
  <el-upload
    v-bind="$attrs"
    :action="uploadUrl"
    :headers="headers"
    :before-upload="beforeUpload"
    :on-success="onSuccess"
    :on-error="onError"
  >
    <slot>
      <el-button type="primary">
        <el-icon><Upload /></el-icon>
        上传文件
      </el-button>
    </slot>
    <template #tip>
      <div v-if="tip" class="el-upload__tip">
        {{ tip }}
      </div>
    </template>
  </el-upload>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  uploadUrl?: string
  maxSize?: number // MB
  accept?: string
  tip?: string
  authorized?: boolean
}>(), {
  maxSize: 10,
  accept: '.pdf,.doc,.docx,.xls,.xlsx,.jpg,.png',
  authorized: true,
})

const emit = defineEmits<{
  success: [response: unknown]
  error: [err: Error]
}>()

function getToken(): string {
  return sessionStorage.getItem('token') || localStorage.getItem('token') || ''
}

const headers = computed(() => ({
  Authorization: `Bearer ${getToken()}`,
}))

function beforeUpload(file: File): boolean | Promise<boolean> {
  const isLtSize = file.size / 1024 / 1024 < props.maxSize
  if (!isLtSize) {
    ElMessage.error(`文件大小不能超过 ${props.maxSize}MB`)
    return false
  }
  return true
}

function onSuccess(response: unknown) {
  emit('success', response)
  ElMessage.success('上传成功')
}

function onError(err: Error) {
  emit('error', err)
  ElMessage.error('上传失败')
}
</script>