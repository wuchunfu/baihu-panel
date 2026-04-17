<script setup lang="ts">
import { ref, watch } from 'vue'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Search, RefreshCw, Trash2, ShieldAlert, Bell, LogIn } from 'lucide-vue-next'
import LoginLogTab from './tabs/LoginLogTab.vue'
import SystemEventTab from './tabs/SystemEventTab.vue'
import PushLog from '@/views/notify/components/PushLog.vue'
import { LOG_LEVEL, LOG_STATUS } from '@/api'

const activeTab = ref('system')
const systemTabRef = ref()
const pushLogRef = ref()
const loginTabRef = ref()

const filters = ref({
  system: { keyword: '', level: 'all' },
  push: { keyword: '', status: 'all' },
  login: { username: '' }
})

let searchTimer: ReturnType<typeof setTimeout> | null = null

function handleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    handleRefresh()
  }, 300)
}

function handleRefresh() {
  if (activeTab.value === 'system') systemTabRef.value?.fetchLogs()
  else if (activeTab.value === 'push') pushLogRef.value?.fetchLogs()
  else if (activeTab.value === 'login') loginTabRef.value?.loadLogs()
}

function handleClear() {
  if (activeTab.value === 'system' && systemTabRef.value) systemTabRef.value.showClearConfirm = true
  else if (activeTab.value === 'push' && pushLogRef.value) pushLogRef.value.showClearConfirm = true
}

// 切换标签时重置搜索
watch(activeTab, () => {
  // handleRefresh()
})

</script>

<template>
  <Tabs v-model="activeTab" class="space-y-6 h-full flex flex-col">
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4 shrink-0 px-1">
      <div class="flex flex-col shrink-0">
        <h2 class="text-xl sm:text-2xl font-bold tracking-tight">消息日志</h2>
        <p class="text-muted-foreground text-sm">
          {{ activeTab === 'system' ? '查看系统重要运行事件' :
            activeTab === 'push' ? '查看消息推送历史记录' : '查看系统用户登录记录' }}
        </p>
      </div>

      <div :class="[activeTab === 'login' ? 'flex flex-row lg:flex-row' : 'flex flex-col lg:flex-row', 'lg:items-center gap-2 lg:gap-3 w-full lg:w-auto lg:ml-auto lg:justify-end']">
        <!-- 搜索与筛选区域 -->
        <div :class="[activeTab === 'login' ? 'flex-1 min-w-0' : 'w-full lg:w-auto', 'flex items-center gap-2']">
          <!-- 系统事件 / 推送日志 搜索框 -->
          <div v-if="activeTab !== 'login'" class="relative flex-1 lg:w-60 group">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
            <Input 
              v-if="activeTab === 'system'"
              v-model="filters.system.keyword" 
              placeholder="搜索系统事件..." 
              class="h-9 pl-9 w-full bg-muted/20 border-muted-foreground/10 focus:bg-background text-sm"
              @input="handleSearch" 
            />
            <Input 
              v-else-if="activeTab === 'push'"
              v-model="filters.push.keyword" 
              placeholder="搜索推送日志..." 
              class="h-9 pl-9 w-full bg-muted/20 border-muted-foreground/10 focus:bg-background text-sm"
              @input="handleSearch" 
            />
          </div>
          <!-- 登录日志 搜索框 -->
          <div v-else class="relative flex-1 lg:w-48 group">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
            <Input 
              v-model="filters.login.username" 
              placeholder="搜索..." 
              class="h-9 pl-9 w-full bg-muted/20 border-muted-foreground/10 focus:bg-background text-sm"
              @input="handleSearch" 
            />
          </div>

          <!-- 系统事件 级别筛选 -->
          <div v-if="activeTab === 'system'" class="relative w-24 sm:w-28 shrink-0">
            <Select v-model="filters.system.level" @update:model-value="handleRefresh">
              <SelectTrigger class="h-9 w-full text-sm bg-muted/20 border-muted-foreground/10">
                <SelectValue placeholder="级别" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">所有级别</SelectItem>
                <SelectItem :value="LOG_LEVEL.INFO">信息</SelectItem>
                <SelectItem :value="LOG_LEVEL.WARNING">警告</SelectItem>
                <SelectItem :value="LOG_LEVEL.ERROR">错误</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- 推送日志 状态筛选 -->
          <div v-if="activeTab === 'push'" class="relative w-24 sm:w-28 shrink-0">
            <Select v-model="filters.push.status" @update:model-value="handleRefresh">
              <SelectTrigger class="h-9 w-full text-sm bg-muted/20 border-muted-foreground/10">
                <SelectValue placeholder="状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">所有状态</SelectItem>
                <SelectItem :value="LOG_STATUS.SUCCESS">发送成功</SelectItem>
                <SelectItem :value="LOG_STATUS.FAILED">发送失败</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <!-- 操作区域 -->
        <div :class="[activeTab === 'login' ? 'flex-1 sm:flex-none' : 'w-full lg:w-auto', 'flex items-center gap-2']">
          <Button variant="outline" size="icon" class="h-9 w-9 shrink-0 shadow-sm" @click="handleRefresh" title="刷新">
            <RefreshCw class="h-4 w-4" />
          </Button>

          <Button v-if="activeTab !== 'login'" variant="outline" 
            class="flex-1 lg:flex-none lg:px-3 h-9 shadow-sm text-destructive border-destructive/20 hover:bg-destructive/10" 
            @click="handleClear" title="清空记录">
            <Trash2 class="h-4 w-4 lg:mr-2" /> <span class="ml-2 lg:inline" :class="activeTab !== 'login' ? '' : 'hidden'">清空记录</span>
          </Button>

          <!-- 桌面端标签切换 -->
          <TabsList class="h-9 p-1 bg-muted/30 border shrink-0 hidden lg:flex">
            <TabsTrigger value="system" class="px-4 h-7 text-sm">
              系统事件
            </TabsTrigger>
            <TabsTrigger value="push" class="px-4 h-7 text-sm">
              推送日志
            </TabsTrigger>
            <TabsTrigger value="login" class="px-4 h-7 text-sm">
              登录日志
            </TabsTrigger>
          </TabsList>

          <!-- 移动端标签切换 (简易版) -->
          <div class="lg:hidden flex-1 shrink-0 min-w-0">
            <Select v-model="activeTab">
              <SelectTrigger class="h-9 w-full text-sm bg-muted/20 border-muted-foreground/10">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="system">系统事件</SelectItem>
                <SelectItem value="push">推送日志</SelectItem>
                <SelectItem value="login">登录日志</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>
    </div>

    <div class="flex-1 min-h-0">
      <TabsContent value="system" class="mt-0">
        <SystemEventTab ref="systemTabRef" :filters="filters.system" />
      </TabsContent>

      <TabsContent value="push" class="mt-0">
        <PushLog ref="pushLogRef" :filters="filters.push" />
      </TabsContent>

      <TabsContent value="login" class="mt-0">
        <LoginLogTab ref="loginTabRef" :username="filters.login.username" />
      </TabsContent>
    </div>
  </Tabs>
</template>
