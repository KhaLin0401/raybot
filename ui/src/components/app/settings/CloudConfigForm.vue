<script setup lang="ts">
import type { CloudConfig } from '@/types/config'
import { useQueryClient } from '@tanstack/vue-query'
import { toTypedSchema } from '@vee-validate/zod'
import { Loader } from 'lucide-vue-next'
import { useForm } from 'vee-validate'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input, PasswordInput } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { CLOUD_CONFIG_QUERY_KEY, useCloudConfigMutation } from '@/composables/use-config'

interface Props {
  initialValues: CloudConfig
}
const props = defineProps<Props>()

const cloudConfigSchema = z.object({
  enable: z.boolean().default(false),
  address: z.string().min(1, 'Address is required'),
  token: z.string().min(1, 'Token is required'),
})

const queryClient = useQueryClient()
const { mutate, isPending } = useCloudConfigMutation()

const form = useForm({
  validationSchema: toTypedSchema(cloudConfigSchema),
  initialValues: props.initialValues,
})

const onSubmit = form.handleSubmit((values) => {
  mutate(values, {
    onSuccess: () => {
      queryClient.setQueryData([CLOUD_CONFIG_QUERY_KEY], values)
      notification.success('Cloud configuration updated successfully!')
    },
    onError: () => {
      notification.error('Failed to update cloud configuration')
    },
  })
})
</script>

<template>
  <form class="flex flex-col w-full max-w-lg space-y-6" @submit="onSubmit">
    <h3 class="pb-2 text-lg font-medium border-b">
      Cloud Configuration
    </h3>

    <FormField v-slot="{ componentField }" name="address">
      <FormItem>
        <FormLabel>Cloud Address</FormLabel>
        <FormControl>
          <Input
            v-bind="componentField"
            type="url"
            placeholder="https://cloud.example.com"
            :disabled="isPending"
          />
        </FormControl>
        <FormMessage />
      </FormItem>
    </FormField>

    <FormField v-slot="{ componentField }" name="token">
      <FormItem>
        <FormLabel>Auth token</FormLabel>
        <FormControl>
          <PasswordInput
            v-bind="componentField"
            placeholder="Enter your auth token"
            :disabled="isPending"
          />
        </FormControl>
        <FormMessage />
      </FormItem>
    </FormField>

    <FormField v-slot="{ value, handleChange }" name="enable">
      <FormItem class="flex flex-row items-center justify-between p-4 border rounded-lg">
        <div class="space-y-0.5">
          <FormLabel>Enable Cloud Service</FormLabel>
        </div>
        <FormControl>
          <Switch
            :model-value="value"
            :disabled="isPending"
            aria-readonly
            @update:model-value="handleChange"
          />
        </FormControl>
      </FormItem>
    </FormField>

    <div>
      <Button type="submit" :disabled="isPending">
        <Loader v-if="isPending" class="w-4 h-4 mr-2 animate-spin" />
        Save
      </Button>
    </div>
  </form>
</template>
