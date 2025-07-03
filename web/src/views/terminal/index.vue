<template>
  <div class="app-container">
    <el-form ref="form" :model="terminal" :inline="true" label-width="120px">
      <el-form-item label="Namespace"> <!-- namespace 输入框 -->
        <el-input v-model="terminal.namespace"/>
      </el-form-item>
      <el-form-item label="Pod"> <!-- pod 名称输入框 -->
        <el-input v-model="terminal.pod"/>
      </el-form-item>
      <el-form-item label="Container"> <!-- 容器名称输入框 -->
        <el-input v-model="terminal.container"/>
      </el-form-item>
      <el-form-item label="Command"> <!-- 命令选择框 -->
        <el-select v-model="terminal.shell" placeholder="bash">
          <el-option label="bash" value="bash"/>
          <el-option label="sh" value="sh"/>
        </el-select>
      </el-form-item>
      <el-form-item> <!-- 提交按钮 -->
        <el-button type="primary" @click="onSubmit">Create</el-button>
      </el-form-item>
      <div id="terminal" />
    </el-form>
  </div>
</template>

<script>
import { Terminal } from 'xterm'
import { common as xtermTheme } from 'xterm-style'
import 'xterm/css/xterm.css'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import axios from 'axios'

export default {
  data() {
    return {
      terminal: {
        namespace: 'default',
        shell: 'bash',
        pod: '',
        container: '',
        sessionid: ''
      },
      inputBuffer: '',
      xterm: null,      // 存储Terminal实例
      ws: null,         // 存储WebSocket实例
      resizeHandler: null // 存储resize事件处理器
    }
  },
  methods: {
    async onSubmit() {
      // 创建终端实例
      this.xterm = new Terminal({
        theme: xtermTheme,
        rendererType: 'canvas',
        convertEol: true,
        cursorBlink: true
      })

      // 加载插件
      const fitAddon = new FitAddon()
      this.xterm.loadAddon(fitAddon)
      this.xterm.loadAddon(new WebLinksAddon())

      // 打开终端并调整大小
      this.xterm.open(document.getElementById('terminal'))
      fitAddon.fit()

      try {
        // 获取sessionid
        const { data } = await axios.get('http://8.156.65.148:8080/terminals', {
          params: {
            namespace: this.terminal.namespace,
            pod_name: this.terminal.pod,
            container_name: this.terminal.container,
            shell: this.terminal.shell
          }
        })

        this.terminal.sessionid = data.id

        // 创建WebSocket连接
        this.ws = new WebSocket(`ws://8.156.65.148:8080/ws/${this.terminal.sessionid}`)

        // 使用箭头函数保持this上下文
        this.ws.onopen = () => {
          this.ws.send(JSON.stringify({
            Op: 'resize',
            Rows: this.xterm.rows,
            Cols: this.xterm.cols
          }))
        }

        // 使用箭头函数保持this上下文
        this.ws.onmessage = (evt) => {
          try {
            const msg = JSON.parse(evt.data)
            if (msg.Op === 'stdout') {
              this.xterm.write(msg.Data)
            } else {
              console.error('Unknown message type:', msg.Op)
            }
          } catch (error) {
            console.error('Error parsing message:', error)
          }
        }

        // 使用箭头函数保持this上下文
        this.ws.onerror = (evt) => {
          this.xterm.write(`WebSocket Error: ${evt.message}\n`)
        }

        // 使用箭头函数保持this上下文
        this.resizeHandler = () => {
          fitAddon.fit()
          this.ws.send(JSON.stringify({
            Op: 'resize',
            Rows: this.xterm.rows,
            Cols: this.xterm.cols
          }))
        }

        window.addEventListener('resize', this.resizeHandler)

        // 处理终端输入
        this.xterm.onData((char) => {
          if (char === '\r') { // 回车
            const dataToSend = this.inputBuffer + '\n'
            this.sendCommand(dataToSend)
            this.inputBuffer = ''
          }
          else if (char === '\x7F') { // 退格
            if (this.inputBuffer.length > 0) {
              this.inputBuffer = this.inputBuffer.slice(0, -1)
              this.xterm.write('\b \b')
            }
          }
          else if (char === '\x03') { // Ctrl+C
            this.xterm.write('^C\n')
            this.inputBuffer = ''
          }
          else { // 普通字符
            this.inputBuffer += char
            this.xterm.write(char)
          }
        })

      } catch (error) {
        console.error('Failed to initialize terminal:', error)
        if (this.xterm) {
          this.xterm.write(`Error: ${error.message}\n`)
        }
      }
    },

    sendCommand(data) {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        console.log('send data: ', data)
        this.ws.send(JSON.stringify({
          Op: 'stdin',
          Data: data
        }))
      }
    }
  },

  // 添加组件销毁时的清理工作
  beforeDestroy() {
    // 关闭WebSocket连接
    if (this.ws && this.ws.readyState !== WebSocket.CLOSED) {
      this.ws.close()
    }

    // 销毁终端实例
    if (this.xterm) {
      this.xterm.dispose()
    }

    // 移除resize事件监听器
    if (this.resizeHandler) {
      window.removeEventListener('resize', this.resizeHandler)
    }
  }
}
</script>

<style scoped>
.line {
  text-align: center;
}
#terminal {
  height: 400px;
  margin-top: 20px;
  border: 1px solid #ccc;
  border-radius: 4px;
  background-color: #000;
  color: #fff;
}
</style>
