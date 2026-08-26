<template>
  <el-dialog v-model="visible" title="审批操作" width="450px">
    <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
      <el-form-item label="审批结果">
        <el-radio-group v-model="form.action">
          <el-radio value="approve">通过</el-radio>
          <el-radio value="reject">驳回</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="审批意见" prop="comment">
        <el-input
          v-model="form.comment"
          type="textarea"
          :rows="4"
          :placeholder="form.action === 'approve' ? '审批通过（可选填意见）' : '请填写驳回原因'"
        />
      </el-form-item>
      <el-form-item v-if="form.action === 'reject'" label="驳回节点">
        <el-select v-model="form.reject_target" placeholder="选择驳回目标节点" style="width:100%">
          <el-option v-for="node in rejectTargets" :key="node.code" :label="node.name" :value="node.code" />
        </el-select>
      </el-form-item>
      <el-form-item label="附件">
        <el-upload
          :auto-upload="false"
          :limit="3"
          :on-change="handleFileChange"
        >
          <el-button size="small">选择文件</el-button>
        </el-upload>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

interface RejectTarget {
  code: string
  name: string
}

const props = withDefaults(defineProps<{
  rejectTargets?: RejectTarget[]
}>(), {
  rejectTargets: () => [],
})

const emit = defineEmits<{
  submit: [data: { action: string; comment: string; reject_target?: string; files: File[] }]
}>()

const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const files = ref<File[]>([])

const form = reactive({
  action: 'approve',
  comment: '',
  reject_target: '',
})

const formRules: FormRules = {
  comment: [
    {
      validator: (_: unknown, value: string, callback: (e?: Error) => void) => {
        if (form.action === 'reject' && !value) {
          callback(new Error('驳回必须填写原因'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

function open() {
  form.action = 'approve'
  form.comment = ''
  form.reject_target = ''
  files.value = []
  visible.value = true
}

function handleFileChange(uploadFile: { raw?: File }) {
  if (uploadFile.raw) {
    files.value.push(uploadFile.raw)
  }
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    emit('submit', {
      action: form.action,
      comment: form.comment,
      reject_target: form.action === 'reject' ? form.reject_target : undefined,
      files: files.value,
    })
    visible.value = false
  } finally {
    submitting.value = false
  }
}

defineExpose({ open })
</script>