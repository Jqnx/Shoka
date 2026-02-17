<script setup lang="ts">
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
  navigationMenuTriggerStyle,
  navigationMenuTriggerStyleActive,
} from "@/components/ui/navigation-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Sun,
  Moon,
  type LucideIcon,
  Heart,
  Settings,
  Library,
  LogOut,
  Panda,
  Search,
} from "lucide-vue-next";

const { signOut, user } = useAuth();
const color = useColorMode();
const route = useRoute();
const changeColorMode = () => {
  if (color.value === "dark") {
    color.preference = "light";
  } else if (color.value === "light") {
    color.preference = "dark";
  }
};

const pages: { title: string; href: string }[] = [
  {
    title: "Home",
    href: "index",
  },
  {
    title: "Archives",
    href: "a",
  },
  {
    title: "Artists",
    href: "artist",
  },
  {
    title: "Tags",
    href: "tag",
  },
  {
    title: "Characters",
    href: "character",
  },
  {
    title: "Parodies",
    href: "parody",
  },
];

const userMenu: { title: string; href: string; icon: LucideIcon }[] = [
  {
    title: "Favorites",
    href: "favorites",
    icon: Heart,
  },
  {
    title: "Settings",
    href: "settings",
    icon: Settings,
  },
];
</script>

<template>
  <div class="flex flex-col gap-4">
    <header class="bg-secondary py-2.5 shadow-sm flex justify-center">
      <NavigationMenu>
        <NavigationMenuList>
          <NavigationMenuItem>
            <NavigationMenuLink as-child class="bg-primary">
              <NuxtLink to="/">
                <Library class="size-6 stroke-primary-foreground" />
              </NuxtLink>
            </NavigationMenuLink>
          </NavigationMenuItem>
        </NavigationMenuList>
        <NavigationMenuList class="lg:px-24 xl:px-32">
          <NavigationMenuItem v-for="(item, index) in pages" :key="index">
            <NavigationMenuLink
              v-if="route.name == item.href"
              as-child
              :class="navigationMenuTriggerStyleActive()"
              class="bg-secondary text-primary"
            >
              <NuxtLink :to="{ name: item.href }">{{ item.title }}</NuxtLink>
            </NavigationMenuLink>
            <NavigationMenuLink
              v-else
              as-child
              :class="navigationMenuTriggerStyle()"
              class="bg-secondary"
            >
              <NuxtLink :to="{ name: item.href }">{{ item.title }}</NuxtLink>
            </NavigationMenuLink>
          </NavigationMenuItem>
        </NavigationMenuList>
        <NavigationMenuList>
          <NavigationMenuItem>
            <NavigationMenuLink>
              <Search class="size-5 stroke-3 stroke-foreground" />
            </NavigationMenuLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavigationMenuTrigger class="bg-secondary">
              <Avatar>
                <!--
                <AvatarImage :src="user?.image" :alt="user?.displayUsername" />
                -->
                <AvatarFallback>
                  <Panda />
                </AvatarFallback>
              </Avatar>
            </NavigationMenuTrigger>
            <NavigationMenuContent>
              <h4 class="text-center pt-1.5 text-md font-semibold">
                {{ user?.displayUsername }}
              </h4>
              <Separator class="my-3" />
              <ul class="grid w-35 gap-1">
                <li v-for="(item, index) in userMenu" :key="index">
                  <NavigationMenuLink as-child>
                    <NuxtLink
                      :to="{ name: item.href }"
                      class="flex flex-row gap-3 items-center"
                    >
                      <component :is="item.icon" />
                      <span>{{ item.title }}</span>
                    </NuxtLink>
                  </NavigationMenuLink>
                </li>
                <li>
                  <NavigationMenuLink
                    class="flex flex-row gap-3 items-center hover:cursor-pointer"
                    @click="changeColorMode"
                  >
                    <Sun v-if="color.preference === 'dark'" />
                    <Moon v-if="color.preference === 'light'" />
                    <span>Theme</span>
                  </NavigationMenuLink>
                </li>
                <li>
                  <NavigationMenuLink
                    class="flex flex-row gap-3 items-center hover:cursor-pointer"
                    @click="() => signOut({ redirectTo: '/login' })"
                  >
                    <LogOut />
                    <span>Sign Out</span>
                  </NavigationMenuLink>
                </li>
              </ul>
            </NavigationMenuContent>
          </NavigationMenuItem>
        </NavigationMenuList>
      </NavigationMenu>
    </header>
    <main class="lg:px-4 xl:px-24 2xl:px-62">
      <slot />
    </main>
  </div>
</template>
