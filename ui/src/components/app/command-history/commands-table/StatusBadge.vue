<script setup lang="ts">
import type { Component } from 'vue'
import type { CommandStatus } from '@/types/command'
import { AlertCircle, CheckCircle, Clock, XCircle } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'

interface Props {
  status: CommandStatus
}
const props = defineProps<Props>()

const STATUS_LABELS: Record<CommandStatus, string> = {
  QUEUED: 'Queued',
  PROCESSING: 'Processing',
  CANCELING: 'Canceling',
  SUCCEEDED: 'Succeeded',
  FAILED: 'Failed',
  CANCELED: 'Canceled',
}

const STATUS_CLASSES: Record<CommandStatus, string> = {
  QUEUED: 'bg-yellow-500/10 text-yellow-500 hover:bg-yellow-500/20',
  PROCESSING: 'bg-blue-500/10 text-blue-500 hover:bg-blue-500/20',
  CANCELING: 'bg-amber-500/10 text-amber-500 hover:bg-amber-500/20',
  SUCCEEDED: 'bg-green-500/10 text-green-500 hover:bg-green-500/20',
  FAILED: 'bg-red-500/10 text-red-500 hover:bg-red-500/20',
  CANCELED: 'bg-gray-500/10 text-gray-500 hover:bg-gray-500/20',
}

const STATUS_ICONS: Record<CommandStatus, Component> = {
  QUEUED: Clock,
  PROCESSING: Clock,
  CANCELING: Clock,
  SUCCEEDED: CheckCircle,
  FAILED: AlertCircle,
  CANCELED: XCircle,
}

const label = STATUS_LABELS[props.status]
const className = STATUS_CLASSES[props.status]
const icon = STATUS_ICONS[props.status]
</script>

<template>
  <Badge :class="className">
    <component :is="icon" class="w-4 h-4 mr-2" />
    {{ label }}
  </Badge>
</template>
