<script setup lang="ts">
import { computed, ref } from 'vue'
import { Folder, File, ChevronRight, ChevronDown, Trash2, Plus, Download } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import type { FileNode } from '@/api'

defineOptions({
  name: 'FileTreeNode'
})

const props = defineProps<{
  node: FileNode
  expandedDirs: Set<string>
  selectedPath: string | null
  depth?: number
}>()

const emit = defineEmits<{
  select: [node: FileNode]
  delete: [path: string]
  'download-file': [path: string]
  create: [parentDir: string]
  move: [oldPath: string, newPath: string]
}>()

const depth = computed(() => props.depth ?? 0)
const isExpanded = computed(() => props.expandedDirs.has(props.node.path))
const isSelected = computed(() => props.selectedPath === props.node.path)
const isDragOver = ref(false)

function handleSelect() {
  emit('select', props.node)
}

function handleDelete(e: Event) {
  e.stopPropagation()
  emit('delete', props.node.path)
}

function handleDownloadClick(e: Event) {
  e.stopPropagation()
  emit('download-file', props.node.path)
}

function handleCreate(e: Event) {
  e.stopPropagation()
  emit('create', props.node.path)
}

function handleDragStart(e: DragEvent) {
  e.dataTransfer?.setData('text/plain', props.node.path)
  e.dataTransfer!.effectAllowed = 'move'
}

function handleDragOver(e: DragEvent) {
  if (!props.node.isDir) return
  e.preventDefault()
  e.dataTransfer!.dropEffect = 'move'
  isDragOver.value = true
}

function handleDragLeave() {
  isDragOver.value = false
}

function handleDrop(e: DragEvent) {
  e.preventDefault()
  isDragOver.value = false
  if (!props.node.isDir) return

  const sourcePath = e.dataTransfer?.getData('text/plain')
  if (!sourcePath || sourcePath === props.node.path) return

  // 不能移动到自己的子目录
  if (props.node.path.startsWith(sourcePath + '/')) return

  const fileName = sourcePath.split('/').pop()
  const newPath = props.node.path ? `${props.node.path}/${fileName}` : fileName

  if (newPath !== sourcePath) {
    emit('move', sourcePath, newPath!)
  }
}
</script>

<template>
  <div>
    <div :class="[
      'flex items-center gap-1 py-0.5 px-1 rounded cursor-pointer text-xs hover:bg-muted group',
      isSelected && 'bg-accent',
      isDragOver && 'bg-blue-500/20 ring-1 ring-blue-500'
    ]" :style="{ paddingLeft: depth * 12 + 4 + 'px' }" draggable="true" @click="handleSelect"
      @dragstart="handleDragStart" @dragover="handleDragOver" @dragleave="handleDragLeave" @drop="handleDrop">
      <template v-if="node.isDir">
        <ChevronDown v-if="isExpanded" class="h-3 w-3 flex-shrink-0" />
        <ChevronRight v-else class="h-3 w-3 flex-shrink-0" />
      </template>
      <span v-else class="w-3" />
      <Folder v-if="node.isDir" class="h-3 w-3 text-yellow-500 flex-shrink-0" />
      <File v-else class="h-3 w-3 text-blue-500 flex-shrink-0" />
      <span class="truncate flex-1">{{ node.name }}</span>
      <Button v-if="node.isDir" variant="ghost" size="icon" class="h-5 w-5 opacity-0 group-hover:opacity-100"
        @click="handleCreate">
        <Plus class="h-3 w-3" />
      </Button>
      <Button v-if="!node.isDir" variant="ghost" size="icon" class="h-5 w-5 opacity-0 group-hover:opacity-100"
        @click="handleDownloadClick">
        <Download class="h-3 w-3" />
      </Button>
      <Button variant="ghost" size="icon" class="h-5 w-5 opacity-0 group-hover:opacity-100" @click="handleDelete">
        <Trash2 class="h-3 w-3 text-destructive" />
      </Button>
    </div>
    <template v-if="node.isDir && isExpanded && node.children">
      <FileTreeNode v-for="child in node.children" :key="child.path" :node="child" :expanded-dirs="expandedDirs"
        :selected-path="selectedPath" :depth="depth + 1" @select="$emit('select', $event)"
        @delete="$emit('delete', $event)" @download-file="$emit('download-file', $event)"
        @create="$emit('create', $event)" @move="(oldPath, newPath) => $emit('move', oldPath, newPath)" />
    </template>
  </div>
</template>
