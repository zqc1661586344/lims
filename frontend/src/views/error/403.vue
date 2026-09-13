<template>
  <div class="forbidden">
    <h1 class="code">403</h1>
    <p class="msg">抱歉，您没有权限访问此页面</p>
    <p v-if="requiredPerm" class="hint">需要权限：{{ requiredPerm }}</p>
    <div class="actions">
      <el-button type="primary" @click="router.push('/dashboard')">返回首页</el-button>
      <el-button @click="router.back()">返回上一页</el-button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
const router = useRouter()
const route = useRoute()
const requiredPerm = computed(() => (route.meta.requiredPerm as string) || '')
import { computed } from 'vue'
</script>
<style scoped>
.forbidden {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
}
.code {
  font-size: 120px;
  font-weight: bold;
  color: #f56c6c;
  margin: 0;
}
.msg {
  font-size: 16px;
  color: #909399;
  margin: 0;
}
.hint {
  font-size: 13px;
  color: #e6a23c;
  background: #fdf6ec;
  padding: 6px 14px;
  border-radius: 4px;
  margin: 0;
}
.actions {
  margin-top: 16px;
  display: flex;
  gap: 12px;
}
</style>