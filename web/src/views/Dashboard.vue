<template>
  <div>
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ overview.machineTotal }}</div>
            <div class="stat-label">主机总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ overview.serviceTotal }}</div>
            <div class="stat-label">服务总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value" style="color: #67c23a">{{ overview.machineOnline }}</div>
            <div class="stat-label">在线主机</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value" style="color: #409eff">{{ overview.serviceRunning }}</div>
            <div class="stat-label">运行中服务</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover">
      <template #header>
        <span>最近主机</span>
      </template>
      <el-table :data="overview.recentMachines" stripe border style="width: 100%">
        <el-table-column prop="name" label="主机名称" />
        <el-table-column prop="ip" label="IP地址" width="160" />
        <el-table-column prop="machineType" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.machineType === 'CLOUD' ? '' : 'success'" size="small">
              {{ row.machineType === 'CLOUD' ? '云服务器' : '局域网' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="serviceCount" label="服务数" width="80" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
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

onMounted(fetchOverview)
</script>

<style scoped>
.stat-card {
  text-align: center;
  padding: 10px 0;
}

.stat-value {
  font-size: 36px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}
</style>