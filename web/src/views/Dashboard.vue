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

    <div style="margin-bottom: 24px; display: flex; gap: 12px">
      <el-button type="primary" @click="handleExport" :loading="exporting">导出配置</el-button>
      <el-button type="success" @click="triggerImport">导入配置</el-button>
      <input ref="fileInput" type="file" accept=".json" style="display: none" @change="handleImport" />
    </div>

    <el-card shadow="never" class="notice-card">
      <template #header>
        <div class="notice-header">
          <span>{{ notice.title || '通知公告' }}</span>
          <el-button v-if="authStore.isAdmin" type="primary" link size="small" @click="showEditNotice = true">编辑</el-button>
        </div>
      </template>
      <div v-if="notice.content" class="notice-content" v-html="formatNotice(notice.content)"></div>
      <div v-else class="notice-empty">暂无通知</div>
    </el-card>

    <el-dialog v-model="showEditNotice" title="编辑通知公告" width="500px">
      <el-form label-width="60px">
        <el-form-item label="标题">
          <el-input v-model="editTitle" placeholder="通知标题" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="editContent" type="textarea" :rows="10" placeholder="通知内容，支持换行" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditNotice = false">取消</el-button>
        <el-button type="primary" @click="saveNotice" :loading="savingNotice">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, onMounted, computed, h, ref } from 'vue'
import { getOverview } from '../api/machine'
import { exportConfig, importConfig } from '../api/config'
import { getNotice, updateNotice } from '../api/notice'
import { useAuthStore } from '../stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'

const authStore = useAuthStore()

const overview = reactive({
  machineTotal: 0,
  serviceTotal: 0,
  machineOnline: 0,
  serviceRunning: 0
})

const notice = reactive({
  title: '',
  content: ''
})

const showEditNotice = ref(false)
const editTitle = ref('')
const editContent = ref('')
const savingNotice = ref(false)

const fetchOverview = async () => {
  try {
    const res = await getOverview()
    Object.assign(overview, res.data)
  } catch {
  }
}

const fetchNotice = async () => {
  try {
    const res = await getNotice()
    if (res.data) {
      notice.title = res.data.title || ''
      notice.content = res.data.content || ''
    }
  } catch {
  }
}

const saveNotice = async () => {
  savingNotice.value = true
  try {
    await updateNotice({ title: editTitle.value, content: editContent.value })
    notice.title = editTitle.value
    notice.content = editContent.value
    showEditNotice.value = false
    ElMessage.success('通知已更新')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingNotice.value = false
  }
}

import { watch } from 'vue'
watch(showEditNotice, (val) => {
  if (val) {
    editTitle.value = notice.title
    editContent.value = notice.content
  }
})

const formatNotice = (content) => {
  if (!content) return ''
  return content.replace(/\n/g, '<br/>')
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
  { label: '运行中Docker服务', value: overview.serviceRunning, color: '#E0A84C', icon: PlayIcon }
])

onMounted(() => {
  fetchOverview()
  fetchNotice()
})

const exporting = ref(false)
const fileInput = ref(null)

const handleExport = async () => {
  exporting.value = true
  try {
    const res = await exportConfig()
    const blob = new Blob([JSON.stringify(res.data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `service-config-${new Date().toISOString().slice(0, 10)}.json`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('配置导出成功')
  } catch {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

const triggerImport = () => {
  fileInput.value?.click()
}

const handleImport = async (e) => {
  const file = e.target.files?.[0]
  if (!file) return
  try {
    await ElMessageBox.confirm('导入配置将创建新记录（不会覆盖现有数据），确定继续？', '导入配置', {
      confirmButtonText: '确定导入',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const text = await file.text()
    const data = JSON.parse(text)
    const res = await importConfig(data)
    ElMessage.success(`导入成功：${res.data.machines} 台主机、${res.data.dockerServices} 个Docker服务、${res.data.otherServices} 个其他服务、${res.data.egressMethods} 条出站方式`)
    fetchOverview()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('导入失败，请检查文件格式')
    }
  }
  e.target.value = ''
}
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

.notice-card {
  border: 1px solid rgba(0, 0, 0, 0.04) !important;
}

.notice-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  font-size: 16px;
}

.notice-content {
  font-size: 14px;
  line-height: 1.8;
  color: #3a3a3a;
  white-space: pre-line;
}

.notice-empty {
  text-align: center;
  color: #b0b0b0;
  padding: 20px 0;
  font-size: 14px;
}
</style>
