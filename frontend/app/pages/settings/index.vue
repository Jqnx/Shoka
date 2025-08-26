<script lang="ts" setup>
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
useHead({
  title: "Settings",
});
</script>

<template>
  <div class="bg-secondary w-9/10 mx-auto px-4 rounded-xl">
    <Accordion type="single" collapsible>
      <AccordionItem value="item-1">
        <AccordionTrigger class="text-md items-center"
          >Source Settings</AccordionTrigger
        >
        <AccordionContent>
          <NHCredentialsForm />
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  </div>
</template>
