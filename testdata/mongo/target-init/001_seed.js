const dbName = "mimic_target";
const targetDb = db.getSiblingDB(dbName);

targetDb.settings.insertMany([
  {
    key: "commission_percentage",
    value: 10,
    type: "number",
    createdAt: new Date("2026-06-01T00:00:00.000Z"),
    updatedAt: new Date("2026-07-15T00:00:00.000Z"),
    lastModifiedBy: "production-admin"
  }
]);

targetDb.roles.insertMany([
  {
    slug: "admin",
    name: "Administrator",
    permissions: ["settings:read", "settings:write"],
    createdAt: new Date("2026-06-01T00:00:00.000Z"),
    updatedAt: new Date("2026-07-15T00:00:00.000Z")
  }
]);

targetDb.settings.createIndex({ key: 1 }, { unique: true, name: "key_unique" });
targetDb.roles.createIndex({ slug: 1 }, { unique: true, name: "slug_unique" });

