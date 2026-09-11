<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useMetaStore } from '@/stores/meta'

const meta = useMetaStore()
const route = useRoute()
onMounted(() => meta.load())

// 详情页时高亮所属一级菜单
const activeMenu = computed(() => {
  if (route.path.startsWith('/lots')) return '/lots'
  if (route.path.startsWith('/machines')) return '/machines'
  return '/samples'
})
</script>

<template>
  <el-container class="app-container">
    <el-header class="app-header">
      <div class="logo">
        <el-icon :size="22"><Box /></el-icon>
        <span>样品与生产追踪系统</span>
      </div>
      <el-menu mode="horizontal" :default-active="activeMenu" router class="app-menu">
        <el-menu-item index="/lots">
          <el-icon><Histogram /></el-icon>
          <span>批次管理</span>
        </el-menu-item>
        <el-menu-item index="/machines">
          <el-icon><Cpu /></el-icon>
          <span>机台看板</span>
        </el-menu-item>
        <el-menu-item index="/samples">
          <el-icon><List /></el-icon>
          <span>样品管理</span>
        </el-menu-item>
      </el-menu>
      <div class="header-right">
        <el-icon :size="18"><User /></el-icon>
        <span>当前操作员</span>
      </div>
    </el-header>

    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<style>
html,
body,
#app {
  margin: 0;
  height: 100%;
}
body {
  font-family:
    'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial,
    sans-serif;
  background: #f5f7fa;
}
.app-container {
  height: 100%;
}
.app-header {
  display: flex;
  align-items: center;
  gap: 32px;
  padding: 0 24px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  height: 56px;
  box-shadow: 0 1px 4px rgb(0 21 41 / 8%);
}
.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 17px;
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
}
.app-menu {
  flex: 1;
  border-bottom: none;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #606266;
  font-size: 13px;
}
.app-main {
  padding: 20px 24px;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  box-sizing: border-box;
}
</style>
