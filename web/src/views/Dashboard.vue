<template>
  <div>
    <el-row :gutter="20" style="margin-bottom: 24px">
      <el-col :span="6" v-for="item in stats" :key="item.label">
        <div class="stat-card">
          <div class="stat-icon" :style="{ color: item.color }">
            <component :is="item.icon" />
          </div>
          <div class="stat-body">
            <div class="stat-value" :style="{ color: item.color }">{{ item.value }}</div>
            <div class="stat-label">{{ item.label }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-card shadow="never" class="recent-card">
      <template #header>
        <div class="recent-header">
          <span>最近主机</span>
          <span class="recent-count">{{ overview.recentMachines.length }} 台</span>
        </div>
      </template>
      <el-table :data="overview.recentMachines" style="width: 100%">
        <el-table-column prop="name" label="主机名称" min-width="140" />
        <el-table-column prop="ip" label="IP地址" width="160" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.machineType === 'CLOUD' ? '' : 'success'" size="small" effect="plain">
              {{ row.machineType === 'CLOUD' ? '云服务器' : '局域网' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="serviceCount" label="服务数" width="80" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="plain">
              {{ row.status === 1 ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, onMounted, computed, h } from 'vue'
import { getOverview } from '../api/machine'

const overview = reactive({
  machineTotal: 0,
  serviceTotal: 0,
  machineOnline: 0,
  serviceRunning: 0,
  recentMachines: []
})

const fetchOverview = async () => {
  try {
    const res = await getOverview()
    Object.assign(overview, res.data)
  } catch {
  }
}

const ServerIcon = () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.5', width: '28', height: '28' }, [
  h('rect', { x: '2', y: '2', width: '20', height: '8', rx: '2', ry: '2' }),
  h('rect', { x: '2', y: '14', width: '20', height: '8', rx: '2', ry: '2' }),
  h('circle', { cx: '6', cy: '6', r: '1', fill: 'currentColor' }),
  h('circle', { cx: '6', cy: '18', r: '1', fill: 'currentColor' })
])

const BoxIcon = () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.5', width: '28', height: '28' }, [
  h('path', { d: 'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z' }),
  h('polyline', { points: '3.27 6.96 12 12.01 20.73 6.96' }),
  h('line', { x1: '12', y1: '22.08', x2: '12', y2: '12' })
])

const CheckIcon = () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.5', width: '28', height: '28' }, [
  h('path', { d: 'M22 11.08V12a10 10 0 1 1-5.93-9.14' }),
  h('polyline', { points: '22 4 12 14.01 9 11.01' })
])

const PlayIcon = () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.5', width: '28', height: '28' }, [
  h('circle', { cx: '12', cy: '12', r: '10' }),
  h('polygon', { points: '10 8 16 12 10 16 10 8', fill: 'currentColor', stroke: 'none' })
])

const stats = computed(() => [
  { label: '主机总数', value: overview.machineTotal, color: '#5B8DEF', icon: ServerIcon },
  { label: '服务总数', value: overview.serviceTotal, color: '#6DA3C7', icon: BoxIcon },
  { label: '在线主机', value: overview.machineOnline, color: '#5FAE7A', icon: CheckIcon },
  { label: '运行中服务', value: overview.serviceRunning, color: '#E0A84C', icon: PlayIcon }
])

onMounted(fetchOverview)
</script>

<style scoped>
.stat-card {
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-radius: 16px;
  border: 1px solid rgba(0, 0, 0, 0.04);
  padding: 22px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: default;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.03);
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.08);
  background: rgba(255, 255, 255, 1);
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.6);
  transition: transform 0.3s ease;
}

.stat-card:hover .stat-icon {
  transform: scale(1.05);
}

.stat-body {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 30px;
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.5px;
}

.stat-label {
  font-size: 13px;
  color: #8E8E92;
  margin-top: 4px;
  letter-spacing: 0.3px;
}

.recent-card {
  border: 1px solid rgba(0, 0, 0, 0.04) !important;
}

.recent-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.recent-count {
  font-size: 13px;
  font-weight: 400;
  color: #999;
}
</style>