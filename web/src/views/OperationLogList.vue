<template>
  <div>
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>操作日志</span>
          <div style="display: flex; gap: 8px">
            <el-select v-model="search.action" placeholder="操作类型" clearable style="width: 120px" @change="fetchData">
              <el-option label="新增" value="create" />
              <el-option label="修改" value="update" />
              <el-option label="删除" value="delete" />
            </el-select>
            <el-select v-model="search.target" placeholder="目标类型" clearable style="width: 140px" @change="fetchData">
              <el-option label="主机" value="machine" />
              <el-option label="Docker服务" value="docker_service" />
              <el-option label="其他服务" value="other_service" />
              <el-option label="出站方式" value="egress_method" />
            </el-select>
            <el-input v-model="search.username" placeholder="操作人" clearable style="width: 120px" @clear="fetchData" @keyup.enter="fetchData" />
          </div>
        </div>
      </template>

      <el-table :data="list" stripe border v-loading="loading">
        <el-table-column label="时间" width="170" align="center">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column prop="username" label="操作人" width="100" align="center" />
        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="actionTag(row.action)" size="small">{{ actionLabel(row.action) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="目标" width="110" align="center">
          <template #default="{ row }">
            {{ targetLabel(row.target) }}
          </template>
        </el-table-column>
        <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        style="margin-top: 16px; justify-content: flex-end"
        @change="fetchData"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getOperationLogs } from '../api/operationLog'

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const search = reactive({
  action: '',
  target: '',
  username: ''
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getOperationLogs({
      page: page.value,
      pageSize: pageSize.value,
      ...search
    })
    list.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const actionLabel = (action) => ({ create: '新增', update: '修改', delete: '删除' }[action] || action)
const actionTag = (action) => ({ create: 'success', update: 'warning', delete: 'danger' }[action] || 'info')
const targetLabel = (target) => ({
  machine: '主机',
  docker_service: 'Docker服务',
  other_service: '其他服务',
  egress_method: '出站方式'
}[target] || target)

const formatTime = (t) => {
  if (!t) return ''
  const d = new Date(t)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

onMounted(fetchData)
</script>
