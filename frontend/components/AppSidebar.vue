<script setup lang="ts">
  import NavMain from "@/components/NavMain.vue";

  import NavUser from "@/components/NavUser.vue";
  import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    type SidebarProps,
  } from "@/components/ui/sidebar";
  import {
    Library,
    Image,
    Images,
    Album,
    BookImage,
    BookUser,
    Users,
    User,
    Tag,
    PersonStanding,
    LibrarySquare,
  } from "lucide-vue-next";

  const props = withDefaults(defineProps<SidebarProps>(), {
    variant: "inset",
  });

  const { data: user } = useAuthState();

  const data = {
    user: {
      name: user.value?.username,
      avatar: "",
    },
    navMain: [
      {
        title: "Archives",
        url: "/a",
        icon: BookImage,
        isActive: true,
      },
      {
        title: "Covers",
        url: "#",
        icon: Image,
      },
      {
        title: "Illustrations",
        url: "#",
        icon: Images,
      },
      {
        title: "Spreads",
        url: "#",
        icon: Album,
      },
      {
        title: "Tankoubons",
        url: "#",
        icon: LibrarySquare,
      },
    ],
    navSecondary: [
      {
        title: "Artists",
        url: "/artist",
        icon: User,
      },
      {
        title: "Groups",
        url: "/group",
        icon: Users,
      },
      {
        title: "Tags",
        url: "/tag",
        icon: Tag,
      },
      {
        title: "Characters",
        url: "/character",
        icon: PersonStanding,
      },
      {
        title: "Parodies",
        url: "/parody",
        icon: BookUser,
      },
    ],
  };
</script>

<template>
  <Sidebar v-bind="props">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" as-child>
            <NuxtLink to="/">
              <div
                class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-sidebar-primary-foreground">
                <Library class="size-4" />
              </div>
              <div class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-medium">Shoka</span>
              </div>
            </NuxtLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <NavMain label="Media" :items="data.navMain" />
      <NavMain label="Metadata" :items="data.navSecondary" />

      <!--
      <NavSecondary :items="data.navSecondary" class="mt-auto" />
      -->
    </SidebarContent>
    <SidebarFooter>
      <NavUser :user="data.user" />
    </SidebarFooter>
  </Sidebar>
</template>
