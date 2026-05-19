<template>
  <div>
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>SSH 终端</span>
          <div style="display: flex; gap: 8px; align-items: center">
            <el-select v-model="selectedMachineId" placeholder="选择主机" style="width: 220px" @change="handleMachineChange">
              <el-option v-for="m in machines" :key="m.id" :label="m.name + ' (' + m.ip + ')'" :value="m.id">
                <span>{{ m.name }}</span>
                <span style="float: right; color: #999; font-size: 12px">{{ m.ip }}</span>
              </el-option>
            </el-select>
            <el-button type="primary" @click="connect" :disabled="!selectedMachineId || connected">连接</el-button>
            <el-button type="danger" @click="disconnect" :disabled="!connected">断开</el-button>
            <el-button :type="showFileManager ? 'warning' : 'success'" @click="toggleFileManager" :disabled="!selectedMachineId">
              {{ showFileManager ? '关闭文件管理' : '文件管理' }}
            </el-button>
          </div>
        </div>
      </template>

      <div :style="{ display: showFileManager ? 'flex' : 'block', gap: '12px' }">
        <div ref="terminalContainer" class="terminal-container" :style="{ flex: showFileManager ? '1 1 60%' : '1', minWidth: 0 }"></div>

        <FileManager
          v-if="showFileManager && selectedMachineId"
          :machine-id="selectedMachineId"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, onActivated, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { getMachines } from '../api/machine'
import { ElMessage } from 'element-plus'
import FileManager from '../components/FileManager.vue'

defineOptions({ name: 'SSHTerminal' })

const machines = ref([])
const selectedMachineId = ref('')
const connected = ref(false)
const terminalContainer = ref(null)
const showFileManager = ref(false)

let term = null
let fitAddon = null
let ws = null
let connectionId = 0

const fetchMachines = async () => {
  try {
    const res = await getMachines({ pageSize: 999 })
    machines.value = (res.data.list || []).filter(m => m.sshEnabled)
  } catch {}
}

const initTerminal = () => {
  if (term) {
    term.dispose()
    term = null
  }
  term = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: '"Cascadia Code", "Fira Code", Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#1e1e2e',
      foreground: '#cdd6f4',
      cursor: '#f5e0dc',
      selectionBackground: '#585b70',
      black: '#45475a',
      red: '#f38ba8',
      green: '#a6e3a1',
      yellow: '#f9e2af',
      blue: '#89b4fa',
      magenta: '#f5c2e7',
      cyan: '#94e2d5',
      white: '#bac2de',
      brightBlack: '#585b70',
      brightRed: '#f38ba8',
      brightGreen: '#a6e3a1',
      brightYellow: '#f9e2af',
      brightBlue: '#89b4fa',
      brightMagenta: '#f5c2e7',
      brightCyan: '#94e2d5',
      brightWhite: '#a6adc8'
    }
  })
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalContainer.value)
  fitAddon.fit()

  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(new Uint8Array([0, ...new TextEncoder().encode(data)]))
    }
  })

  term.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(new Uint8Array([1, rows, cols]))
    }
  })

  window.addEventListener('resize', handleResize)
}

const handleResize = () => {
  if (fitAddon) fitAddon.fit()
}

const closeWs = () => {
  if (ws) {
    ws.onopen = null
    ws.onmessage = null
    ws.onclose = null
    ws.onerror = null
    if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
      ws.close()
    }
    ws = null
  }
  connected.value = false
}

const connect = async () => {
  if (!selectedMachineId.value) return

  closeWs()

  connectionId++
  const currentConnId = connectionId

  if (term) term.reset()

  const token = localStorage.getItem('token')
  if (!token) {
    ElMessage.error('未登录，请先登录')
    return
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const cols = term ? term.cols : 120
  const rows = term ? term.rows : 30
  const url = `${protocol}//${host}/api/ssh-terminal/${selectedMachineId.value}?token=${encodeURIComponent(token)}&cols=${cols}&rows=${rows}`

  term.writeln('\r\n\x1b[33m正在连接...\x1b[0m')

  try {
    ws = new WebSocket(url)
  } catch (e) {
    term.writeln('\r\n\x1b[31mWebSocket创建失败: ' + e.message + '\x1b[0m')
    return
  }
  ws.binaryType = 'arraybuffer'

  ws.onopen = () => {
    if (currentConnId !== connectionId) return
    connected.value = true
    term.writeln('\x1b[32m已连接\x1b[0m\r\n')
  }

  ws.onmessage = (event) => {
    if (currentConnId !== connectionId) return
    if (event.data instanceof ArrayBuffer) {
      term.write(new Uint8Array(event.data))
    } else {
      term.write(event.data)
    }
  }

  ws.onclose = (event) => {
    if (currentConnId !== connectionId) return
    connected.value = false
    if (event.code === 1006) {
      term.writeln('\r\n\x1b[31m连接异常断开（可能是认证失败或服务端未启动）\x1b[0m')
    } else {
      term.writeln('\r\n\x1b[31m连接已断开 (code: ' + event.code + ')\x1b[0m')
    }
  }

  ws.onerror = () => {
    if (currentConnId !== connectionId) return
    connected.value = false
    term.writeln('\r\n\x1b[31m连接出错\x1b[0m')
    ElMessage.error('WebSocket连接失败，请检查后端服务是否启动')
  }
}

const disconnect = () => {
  closeWs()
  if (term) {
    term.writeln('\r\n\x1b[33m已断开连接\x1b[0m')
  }
}

const handleMachineChange = () => {
  closeWs()
  if (term) term.reset()
}

const toggleFileManager = () => {
  showFileManager.value = !showFileManager.value
  nextTick(() => {
    if (fitAddon) fitAddon.fit()
  })
}

onMounted(async () => {
  await nextTick()
  initTerminal()
  fetchMachines()
})

onActivated(() => {
  nextTick(() => {
    if (fitAddon && term) {
      fitAddon.fit()
    }
  })
})

onBeforeUnmount(() => {
  closeWs()
  if (term) {
    term.dispose()
    term = null
  }
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: calc(100vh - 220px);
  min-height: 400px;
  background: #1e1e2e;
  border-radius: 8px;
  padding: 4px;
}
</style>
