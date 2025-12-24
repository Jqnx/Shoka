export default defineNuxtRouteMiddleware(async (to) => {
  const { fetchSession } = useAuth();
  const data = await fetchSession();
  if (to.path === "/login" || to.path.startsWith("/api")) {
    return;
  }

  if (!data) {
    return navigateTo("/login");
  }
});
