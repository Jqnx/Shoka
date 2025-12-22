import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { db } from "../db";
import { jwt, username } from "better-auth/plugins";
import * as schema from "../db/schema/better-auth";

export const auth = betterAuth({
  database: drizzleAdapter(db, {
    provider: "pg",
    schema: {
      ...schema,
    },
  }),
  emailAndPassword: {
    enabled: true,
  },
  plugins: [jwt(), username()],
  advanced: {
    database: {
      generateId: "uuid",
    },
  },
});
