const dbName = "mimic_source";
const sourceDb = db.getSiblingDB(dbName);

sourceDb.settings.insertMany([
  {
    key: "commission_percentage",
    value: 12,
    type: "number",
    createdAt: new Date("2026-07-01T00:00:00.000Z"),
    updatedAt: new Date("2026-07-29T00:00:00.000Z"),
    lastModifiedBy: "staging-admin"
  },
  {
    key: "dynamic_pricing_enabled",
    value: true,
    type: "boolean",
    createdAt: new Date("2026-07-20T00:00:00.000Z"),
    updatedAt: new Date("2026-07-29T00:00:00.000Z")
  }
]);

sourceDb.roles.insertMany([
  {
    slug: "admin",
    name: "Administrator",
    permissions: ["settings:read", "settings:write", "roles:write"],
    createdAt: new Date("2026-07-01T00:00:00.000Z"),
    updatedAt: new Date("2026-07-29T00:00:00.000Z")
  },
  {
    slug: "dispatcher",
    name: "Dispatcher",
    permissions: ["trips:read", "trips:assign"],
    createdAt: new Date("2026-07-20T00:00:00.000Z"),
    updatedAt: new Date("2026-07-29T00:00:00.000Z")
  }
]);

sourceDb.settings.createIndex({ key: 1 }, { unique: true, name: "key_unique" });
sourceDb.roles.createIndex({ slug: 1 }, { unique: true, name: "slug_unique" });

