import type {
  BetterAuthClientOptions,
  InferSessionFromClient,
  InferUserFromClient,
} from "better-auth/client";
import { createAuthClient } from "better-auth/vue";
import { jwtClient, usernameClient } from "better-auth/client/plugins";
import type { RouteLocationRaw } from "vue-router";

export function useAuth() {
  const url = useRequestURL();
  const headers = import.meta.server ? useRequestHeaders() : undefined;
  const client = createAuthClient({
    baseURL: url.origin,
    fetchOptions: {
      headers,
    },
    plugins: [usernameClient(), jwtClient()],
  });

  const session =
    useState<InferSessionFromClient<BetterAuthClientOptions> | null>(
      "auth:session",
      () => null,
    );
  const user = useState<InferUserFromClient<BetterAuthClientOptions> | null>(
    "auth:user",
    () => null,
  );
  const sessionFetching = import.meta.server
    ? ref(false)
    : useState("auth:sessionFetching", () => false);

  const fetchSession = async () => {
    if (sessionFetching.value) return;
    sessionFetching.value = true;
    try {
      const { data } = await client.getSession({
        fetchOptions: {
          headers,
        },
      });
      session.value = data?.session || null;
      user.value = data?.user || null;
      return data;
    } catch (error) {
      console.error("Error fetching session:", error);
      session.value = null;
      user.value = null;
    } finally {
      sessionFetching.value = false;
    }
  };

  if (import.meta.client) {
    client.$store.listen("$sessionSignal", async (signal) => {
      if (!signal) return;
      await fetchSession();
    });
  }

  const getToken = async () => {
    const { data: token } = await useFetch("/api/get-token", { key: "token" });
    return token.value?.token;
  };

  return {
    session,
    user,
    loggedIn: computed(() => !!session.value),
    signIn: client.signIn,
    signUp: client.signUp,
    resetPassword: client.resetPassword,
    sendVerificationEmail: client.sendVerificationEmail,
    errorCodes: client.$ERROR_CODES,
    async signOut({ redirectTo }: { redirectTo?: RouteLocationRaw } = {}) {
      await client.signOut({
        fetchOptions: {
          onSuccess: async () => {
            session.value = null;
            user.value = null;
            if (redirectTo) {
              await navigateTo(redirectTo);
            }
          },
        },
      });
    },
    client,
    fetchSession,
    getToken,
  };
}
