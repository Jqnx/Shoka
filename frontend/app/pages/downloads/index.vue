<script lang="ts" setup>
import { toast } from "vue-sonner";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from "@/components/ui/form";
import z from "zod";
import { Plus } from "lucide-vue-next";
import type { Download } from "~/types/types";

const wsUrl = useRuntimeConfig().public.ws;
const downloads = ref<Download[]>([]);
const delFile = ref(true);

const { data: initialDownloads, pending } = await useAsyncData(
  "downloads",
  () => $fetch("/api/download"),
);

if (initialDownloads.value) {
  downloads.value = initialDownloads.value;
}

const { data, open } = useWebSocket(`${wsUrl}`, {
  autoReconnect: {
    retries: 3,
    delay: 3000,
  },
  heartbeat: {
    message: "ping",
    interval: 30000,
  },
  autoClose: true,
  onConnected() {
    console.log("Succesfully connected to websocket!");
  },
});

watch(data, async (newData) => {
  if (!newData) return;

  const message = JSON.parse(newData);
  if (message.type === "download_update") {
    console.log(message)
    const updatedDownload = message.data;
    const index = downloads.value.findIndex((d) => d.id === updatedDownload.id);

    if (index >= 0) {
      downloads.value[index] = updatedDownload;
    } else {
      downloads.value.push(updatedDownload);
    }
  } else if (message.type === "download_deleted") {
    const data = message.data;
    const id = data.id;
    downloads.value = downloads.value.filter((d) => d.id != id);
  }
});

const compDownloads = () => {
  return downloads.value.length != 0
    ? `Downloads (${downloads.value.length})`
    : `Downloads`;
};

useHead({
  title: compDownloads(),
});

const formSchema = toTypedSchema(
  z.object({
    url: z.string().url().nonempty(),
  }),
);

const { handleSubmit, errors } = useForm({
  validationSchema: formSchema,
});

// TODO: Update to support failed downloads (url not supported)
const addDownload = handleSubmit((values) => {
  $fetch("/api/download", {
    method: "post",
    body: JSON.stringify(values, null, 2),
    onResponseError({ response }) {
      const err = JSON.stringify(response._data.data, null, 2);
      const test = JSON.parse(err);
      toast.error(h("pre", test.url));
    },
    onResponse() {
      toast.success(h("pre", "Added download."));
    },
  });
});

const deleteDownload = async (id: string) => {
  $fetch(`/api/download/${id}`, {
    method: "delete",
    query: {
      file: delFile.value,
    },
    onResponseError({ response }) {
      const err = JSON.stringify(response._data.message, null, 2);
      toast.error(JSON.parse(err));
    },
    onResponse({ response }) {
      if (response.ok) {
        toast.success(h("pre", "Removed download."));
      }
    },
  });
};

for (let index = 0; index < downloads.value.length; index++) {
  const dl = downloads.value[index];
  console.log(dl);
}

onMounted(() => {
  open();
});
</script>

<template>
  <div class="flex flex-1 flex-col items-center gap-4 p-4">
    <form class="flex justify-center w-full max-w-xl" @submit="addDownload">
      <FormField v-slot="{ componentField }" name="url">
        <FormItem class="w-full">
          <FormControl>
            <Input
              class="rounded-r-none w-full"
              type="url"
              placeholder="Add download url"
              v-bind="componentField"
            />
          </FormControl>
          <FormLabel v-if="errors.url">
            <p class="text-destructive">
              {{ errors.url }}
            </p>
          </FormLabel>
        </FormItem>
      </FormField>
      <Button class="rounded-l-none stroke-primary-foreground" type="submit">
        <Plus />
      </Button>
    </form>
    <div class="flex flex-col items-center w-full gap-2">
      <h1 class="text-2xl font-semibold py-4">Downloads Queue</h1>

      <div
        v-if="!pending && downloads.length === 0"
        class="p-6 text-center text-foreground/50"
      >
        No downloads yet.
      </div>

      <div
        v-else-if="!pending"
        class="flex flex-col items-center w-full gap-2 divide-y divide-gray-200"
      >
        <DownloadItem
          v-for="download in downloads"
          :key="download.id"
          :download="download"
          @delete="deleteDownload(download.id)"
        />
      </div>
    </div>
  </div>
</template>
