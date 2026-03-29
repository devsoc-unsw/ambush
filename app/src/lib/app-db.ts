/**
 * Drizzle DB — server-only (SSR). Imports from `app/db`.
 */
import { createDb, schema } from "ambush/db";

export const db = createDb();
export { schema };
export type { Schema } from "ambush/db";
