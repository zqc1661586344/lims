<template>
  <div class="not-found">
    <h1 class="code">404</h1>
    <p class="msg">抱歉，您访问的页面不存在</p>
    <el-button type="primary" @click="router.push('/dashboard')">返回首页</el-button>
  </div>
</template>
<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>
<style scoped>
.not-found {
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
  color: #409eff;
  margin: 0;
}
.msg {
  font-size: 16px;
  color: #909399;
  margin: 0 0 16px;
}
</style>