<template>
  <div class="app-container">
    <!-- 使用 Element UI 的表单组件创建一个带有标签和输入框的表单 -->
    <el-form ref="form" :model="terminal" :inline="true" label-width="120px">
      <el-form-item label="Namespace"> <!-- namespace 输入框 -->
        <el-input v-model="terminal.namespace" />
      </el-form-item>
      <el-form-item label="Pod"> <!-- pod 名称输入框 -->
        <el-input v-model="terminal.pod" />
      </el-form-item>
      <el-form-item label="Container"> <!-- 容器名称输入框 -->
        <el-input v-model="terminal.container" />
      </el-form-item>
      <el-form-item label="Command"> <!-- 命令选择框 -->
        <el-select v-model="terminal.shell" placeholder="bash">
          <el-option label="bash" value="bash" />
          <el-option label="sh" value="sh" />
        </el-select>
      </el-form-item>
      <el-form-item> <!-- 提交按钮 -->
        <el-button type="primary" @click="onSubmit">Create</el-button>
      </el-form-item>
      <div id="terminal" /> <!-- 终端视图容器 -->
    </el-form>
  </div>
</template>

<script>
import { Terminal } from 'xterm' // 导入 xterm 包，用于创建和操作终端对象
import { common as xtermTheme } from 'xterm-style' // 导入 xterm 样式主题
import 'xterm/css/xterm.css' // 导入 xterm CSS 样式
import { FitAddon } from 'xterm-addon-fit' // 导入 xterm fit 插件，用于调整终端大小
import { WebLinksAddon } from 'xterm-addon-web-links' // 导入 xterm web-links 插件，可以捕获 URL 并将其转换为可点击链接
import 'xterm/lib/xterm.js'
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
      inputBuffer: ''
    }
  },
  methods: {
    async onSubmit() {
      // 创建一个新的 Terminal 对象
      const xterm = new Terminal({
        theme: xtermTheme,
        rendererType: 'canvas',
        convertEol: true,
        cursorBlink: true
      })

      // 创建并加载 FitAddon 和 WebLinksAddon
      const fitAddon = new FitAddon()
      xterm.loadAddon(fitAddon)
      xterm.loadAddon(new WebLinksAddon())

      // 打开这个终端，并附加到 HTML 元素上
      xterm.open(document.getElementById('terminal'))

      // 调整终端的大小以适应其父元素
      fitAddon.fit()
      console.log('get session id')
      // 获取sessionid
      const data = await axios.get('http://8.156.65.148:8080/terminals', {
        params: {
          namespace: this.terminal.namespace,
          pod_name: this.terminal.pod,
          container_name: this.terminal.container,
          shell: this.terminal.shell
        }
      })
      console.log('data is', data)
      const id = data.id
      console.log('sessionid is', id)
      console.log('new websocket')
      // 创建一个新的 WebSocket 连接，并通过 URL 参数传递 pod, namespace, container 和 command 信息
      const ws = new WebSocket(`ws://8.156.65.148:8080/ws/${id}`)

      // 当 WebSocket 连接打开时，发送一个 resize 消息给服务器，告诉它终端的尺寸
      ws.onopen = function() {
        ws.send(JSON.stringify({
          Op: 'resize',
          Rows: xterm.rows,
          Cols: xterm.cols
        }))
      }

      // 当从服务器收到消息时，写入终端显示
      ws.onmessage = function(evt) {
        try {
          const msg = JSON.parse(evt.data)
          if (msg.Op === 'stdout') {
            xterm.write(msg.Data)
          } else {
            console.error('Unknown message type:', msg.Op)
          }
        } catch (error) {
          console.error('Error parsing message:', error)
        }
      }

      // 当发生错误时，也写入终端显示
      ws.onerror = function(evt) {
        xterm.write(evt.data)
      }

      // 当窗口尺寸变化时，重新调整终端的尺寸，并发送一个新的 resize 消息给服务器
      window.addEventListener('resize', function() {
        fitAddon.fit()
        ws.send(JSON.stringify({
          Op: 'resize',
          Rows: xterm.rows,
          Cols: xterm.cols
        }))
      })

      // 当在终端中键入字符时，发送消息给服务器
      xterm.onData((char) => {
        // 处理回车符
        if (char === '\r') {
          // 无论缓冲区是否有内容，都发送至少一个换行符
          const dataToSend = this.inputBuffer || '\n';
          this.sendCommand(dataToSend);

          // 清空缓冲区并显示换行
          this.inputBuffer = '';
          this.xterm.write('\n');

          // 显示新的命令提示符
          this.xterm.write(this.prompt);
        }
        // 处理退格键
        else if (char === '\x7F') {
          if (this.inputBuffer.length > 0) {
            // 删除缓冲区最后一个字符
            this.inputBuffer = this.inputBuffer.slice(0, -1);
            // 在终端中模拟退格效果
            this.xterm.write('\b \b');
          }
        }
        // 处理Ctrl+C（中断当前命令）
        else if (char === '\x03') {
          this.xterm.write('^C\n');
          this.inputBuffer = '';
          this.xterm.write(this.prompt);
        }
        // 处理Ctrl+D（EOF，通常用于退出shell）
        else if (char === '\x04') {
          if (this.inputBuffer.length === 0) {
            // 如果缓冲区为空，发送EOF并关闭连接
            this.xterm.write('^D\n');
            this.sendCommand('\x04');
            // 可以选择在这里关闭WebSocket连接
          } else {
            // 如果缓冲区有内容，将Ctrl+D作为普通字符处理
            this.inputBuffer += char;
            this.xterm.write(char);
          }
        }
        // 处理其他普通字符
        else {
          this.inputBuffer += char;
          this.xterm.write(char);
        }
      });
    },
    sendCommand(data) {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({
          Op: 'stdin',
          Data: data
        }));
      }
    }
  }
}
</script>

<style scoped>
.line{
  text-align: center;
}
</style>
