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
          <span>通知公告</span>
          <el-button v-if="authStore.isAdmin" type="primary" link size="small" @click="openAddNotice">新增通知</el-button>
        </div>
      </template>
      <div v-if="notices.length > 0" class="notice-list">
        <div v-for="item in notices" :key="item.id" class="notice-item">
          <div class="notice-item-header">
            <span class="notice-item-title">{{ item.title }}</span>
            <div class="notice-item-actions" v-if="authStore.isAdmin">
              <el-button type="primary" link size="small" @click="openEditNotice(item)">编辑</el-button>
              <el-button type="danger" link size="small" @click="handleDeleteNotice(item)">删除</el-button>
            </div>
          </div>
          <div class="notice-content markdown-body" v-html="renderMarkdown(item.content)"></div>
        </div>
      </div>
      <div v-else class="notice-empty">暂无通知</div>
    </el-card>

    <el-dialog v-model="showEditNotice" :title="editingNoticeId ? '编辑通知' : '新增通知'" width="600px">
      <el-form label-width="60px">
        <el-form-item label="标题">
          <el-input v-model="editTitle" placeholder="通知标题" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="editContent" type="textarea" :rows="12" placeholder="通知内容，支持 Markdown 语法" />
        </el-form-item>
        <el-form-item label="预览">
          <div class="notice-preview markdown-body" v-html="renderMarkdown(editContent)"></div>
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
import { getNotices, createNotice, updateNotice, deleteNotice } from '../api/notice'
import { useAuthStore } from '../stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import { marked } from 'marked'

const authStore = useAuthStore()

const overview = reactive({
  machineTotal: 0,
  serviceTotal: 0,
  machineOnline: 0,
  dockerRunning: 0,
  otherRunning: 0
})

const notices = ref([])
const showEditNotice = ref(false)
const editingNoticeId = ref(null)
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

const fetchNotices = async () => {
  try {
    const res = await getNotices()
    notices.value = res.data || []
  } catch {
  }
}

const renderMarkdown = (content) => {
  if (!content) return ''
  return marked(content, { breaks: true })
}

const openAddNotice = () => {
  editingNoticeId.value = null
  editTitle.value = ''
  editContent.value = ''
  showEditNotice.value = true
}

const openEditNotice = (item) => {
  editingNoticeId.value = item.id
  editTitle.value = item.title
  editContent.value = item.content
  showEditNotice.value = true
}

const saveNotice = async () => {
  if (!editTitle.value.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  savingNotice.value = true
  try {
    if (editingNoticeId.value) {
      await updateNotice(editingNoticeId.value, { title: editTitle.value, content: editContent.value })
    } else {
      await createNotice({ title: editTitle.value, content: editContent.value })
    }
    showEditNotice.value = false
    ElMessage.success(editingNoticeId.value ? '通知已更新' : '通知已创建')
    fetchNotices()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingNotice.value = false
  }
}

const handleDeleteNotice = async (item) => {
  try {
    await ElMessageBox.confirm(`确定删除通知「${item.title}」吗？`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteNotice(item.id)
    ElMessage.success('已删除')
    fetchNotices()
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

const PlayIcon = () => h('svg', { viewBox: '0 0 1024 1024', width: '28', height: '28' }, [
  h('path', { d: 'M512 937.353846c-234.732308 0-425.353846-190.621538-425.353846-425.353846 0-17.329231 14.178462-31.507692 31.507692-31.507692h236.307692c17.329231 0 31.507692 14.178462 31.507693 31.507692 0 69.316923 56.713846 126.030769 126.030769 126.030769s126.030769-56.713846 126.030769-126.030769c0-17.329231 14.178462-31.507692 31.507693-31.507692h236.307692c17.329231 0 31.507692 14.178462 31.507692 31.507692 0 234.732308-190.621538 425.353846-425.353846 425.353846zM151.236923 543.507692c15.753846 185.107692 171.716923 330.830769 360.763077 330.83077s345.009231-145.723077 360.763077-330.83077H698.683077C683.716923 632.516923 605.735385 701.046154 512 701.046154s-171.716923-68.529231-186.683077-157.538462H151.236923z', fill: 'currentColor' }),
  h('path', { d: 'M512 118.153846c-217.403077 0-393.846154 176.443077-393.846154 393.846154h236.307692c0-86.646154 70.892308-157.538462 157.538462-157.538462s157.538462 70.892308 157.538462 157.538462h236.307692c0-217.403077-176.443077-393.846154-393.846154-393.846154z', fill: '#d60909' }),
  h('path', { d: 'M905.846154 543.507692H669.538462c-17.329231 0-31.507692-14.178462-31.507693-31.507692 0-69.316923-56.713846-126.030769-126.030769-126.030769s-126.030769 56.713846-126.030769 126.030769c0 17.329231-14.178462 31.507692-31.507693 31.507692H118.153846c-17.329231 0-31.507692-14.178462-31.507692-31.507692 0-234.732308 190.621538-425.353846 425.353846-425.353846s425.353846 190.621538 425.353846 425.353846c0 17.329231-14.178462 31.507692-31.507692 31.507692z m-207.163077-63.015384h174.867692C857.009231 295.384615 701.046154 149.661538 512 149.661538S166.990769 295.384615 151.236923 480.492308h174.867692C340.283077 391.483077 418.264615 322.953846 512 322.953846s171.716923 68.529231 186.683077 157.538462z', fill: 'currentColor' }),
  h('path', { d: 'M512 701.046154c-103.975385 0-189.046154-85.070769-189.046154-189.046154s85.070769-189.046154 189.046154-189.046154 189.046154 85.070769 189.046154 189.046154-85.070769 189.046154-189.046154 189.046154z m0-315.076923c-69.316923 0-126.030769 56.713846-126.030769 126.030769s56.713846 126.030769 126.030769 126.030769 126.030769-56.713846 126.030769-126.030769-56.713846-126.030769-126.030769-126.030769z', fill: 'currentColor' }),
  h('path', { d: 'M512 512m-78.769231 0a78.769231 78.769231 0 1 0 157.538462 0 78.769231 78.769231 0 1 0-157.538462 0Z', fill: 'currentColor' })
])

const stats = computed(() => [
  { label: '主机总数', value: overview.machineTotal, color: '#5B8DEF', icon: ServerIcon },
  { label: '服务总数', value: overview.serviceTotal, color: '#6DA3C7', icon: BoxIcon },
  { label: '在线主机', value: overview.machineOnline, color: '#5FAE7A', icon: CheckIcon },
  { label: '运行中Docker服务', value: overview.dockerRunning, color: '#E0A84C', icon: PlayIcon }
])

onMounted(() => {
  fetchOverview()
  fetchNotices()
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

.notice-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.notice-item {
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  padding: 16px;
  background: #fafafa;
}

.notice-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.notice-item-title {
  font-weight: 600;
  font-size: 15px;
  color: #303133;
}

.notice-item-actions {
  display: flex;
  gap: 4px;
}

.notice-content {
  font-size: 14px;
  line-height: 1.8;
  color: #3a3a3a;
}

.notice-preview {
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 12px;
  min-height: 80px;
  background: #fafafa;
  max-height: 300px;
  overflow-y: auto;
}

.notice-empty {
  text-align: center;
  color: #b0b0b0;
  padding: 20px 0;
  font-size: 14px;
}

.markdown-body :deep(h1) { font-size: 1.4em; margin: 0.6em 0 0.4em; font-weight: 700; }
.markdown-body :deep(h2) { font-size: 1.2em; margin: 0.5em 0 0.3em; font-weight: 700; }
.markdown-body :deep(h3) { font-size: 1.1em; margin: 0.4em 0 0.3em; font-weight: 600; }
.markdown-body :deep(p) { margin: 0.4em 0; }
.markdown-body :deep(ul), .markdown-body :deep(ol) { padding-left: 1.5em; margin: 0.4em 0; }
.markdown-body :deep(li) { margin: 0.2em 0; }
.markdown-body :deep(code) { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; font-size: 0.9em; }
.markdown-body :deep(pre) { background: #f5f5f5; padding: 12px; border-radius: 6px; overflow-x: auto; }
.markdown-body :deep(pre code) { background: none; padding: 0; }
.markdown-body :deep(blockquote) { border-left: 4px solid #ddd; margin: 0.5em 0; padding: 0.3em 1em; color: #666; }
.markdown-body :deep(a) { color: #409eff; text-decoration: none; }
.markdown-body :deep(a:hover) { text-decoration: underline; }
.markdown-body :deep(strong) { font-weight: 600; }
.markdown-body :deep(hr) { border: none; border-top: 1px solid #eee; margin: 1em 0; }
.markdown-body :deep(table) { border-collapse: collapse; width: 100%; margin: 0.5em 0; }
.markdown-body :deep(th), .markdown-body :deep(td) { border: 1px solid #ddd; padding: 8px 12px; text-align: left; }
.markdown-body :deep(th) { background: #f5f5f5; font-weight: 600; }
</style>
