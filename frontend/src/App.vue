<template>
  <router-view />
</template>

<script setup lang="ts">
// Root component — router-view renders MainLayout or Login
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html,
body,
#app {
  height: 100%;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Microsoft YaHei',
    Arial, sans-serif;
}

/* 全局卡片头部：标题居左，操作区居右，垂直居中，
   修复部门/用户/角色及各业务节点页面“标题与按钮挤在一起”的问题 */
.el-card__header .card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.el-card__header .card-header > span {
  font-weight: 600;
}

/* 全局表格操作列：按钮横向排列、不换行、保留间距 */
.el-table .cell {
  white-space: nowrap;
}
.el-table__body .cell .el-button + .el-button {
  margin-left: 8px;
}
</style>