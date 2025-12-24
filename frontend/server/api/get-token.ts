import { auth } from "~~/lib/auth";

export default defineEventHandler(async (event) => {
  const token = await auth.api.getToken({
    headers: event.headers,
  });

  return token;
});
