export default defineNuxtRouteMiddleware(async (to) => {
  const { fetchSession } = useAuth();
  const data = await fetchSession();
  if (to.path === "/login") {
    return;
  }

  if (!data) {
    return navigateTo("/login");
  }
});
