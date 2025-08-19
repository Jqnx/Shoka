<script lang="ts" setup>
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Settings } from "lucide-vue-next";
import { VisuallyHidden } from "reka-ui";

const { preload, fit } = storeToRefs(useReaderSettingsStore());
</script>

<template>
  <Dialog>
    <DialogTrigger as-child>
      <Button
        title="Settings"
        class="h-full px-3 rounded-l-none bg-secondary flex items-center justify-center hover:bg-muted-foreground/20 cursor-pointer"
      >
        <Settings class="size-5 stroke-foreground" />
      </Button>
    </DialogTrigger>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Reader Settings</DialogTitle>
        <VisuallyHidden as-child>
          <DialogDescription />
        </VisuallyHidden>
      </DialogHeader>
      <div class="flex flex-col gap-5">
        <div class="flex flex-col gap-3">
          <h1 class="font-bold text-lg">Page Scaling</h1>
          <Button
            v-if="fit === ''"
            variant="default"
            class="cursor-pointer font-normal text-md"
          >
            Natural
          </Button>
          <Button
            v-else
            variant="secondary"
            class="cursor-pointer font-normal text-md"
            @click="fit = ''"
          >
            Natural
          </Button>
          <div class="flex grow items-center justify-between">
            <h1 class="font-medium text-md">Fit</h1>
            <div class="flex grow w-full max-w-4/5">
              <Button
                v-if="fit === 'max-h-screen'"
                class="flex-1 rounded-r-none font-normal text-md cursor-pointer"
                variant="default"
              >
                Fit to screen
              </Button>
              <Button
                v-else
                class="flex-1 rounded-r-none font-normal text-md cursor-pointer"
                variant="secondary"
                @click="fit = 'max-h-screen'"
              >
                Fit to screen
              </Button>
              <Button
                v-if="fit === 'w-full'"
                class="flex-1 rounded-l-none font-normal text-md cursor-pointer"
                variant="default"
              >
                Stretch
              </Button>
              <Button
                v-else
                class="flex-1 rounded-l-none font-normal text-md cursor-pointer"
                variant="secondary"
                @click="fit = 'w-full'"
              >
                Stretch
              </Button>
            </div>
          </div>
          <h1 class="text-lg font-bold">Network</h1>
          <div class="flex flex-col gap-1">
            <h1 class="font-medium text-md">Image Preloading</h1>
            <div class="flex grow">
              <Button
                v-for="n in 6"
                :key="n"
                class="flex-1"
                :class="
                  n === 1
                    ? 'rounded-r-none'
                    : n === 6
                      ? 'rounded-l-none'
                      : 'rounded-none'
                "
                :variant="preload === n ? 'default' : 'secondary'"
                @click="preload = n"
                >{{ n }}</Button
              >
            </div>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
