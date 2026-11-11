const bcrypt = require("/opt/seed/node_modules/bcryptjs");

const databaseName = process.env.MONGODB_DATABASE;
if (!databaseName) throw new Error("MONGODB_DATABASE is required");

const database = db.getSiblingDB(databaseName);
database.admins.createIndex({ username: 1 }, { unique: true });
// Development account: admin / 1234. Never replace an existing password on reseed.
database.admins.updateOne(
  { username: "admin" },
  {
    $setOnInsert: {
      username: "admin",
      password_hash: bcrypt.hashSync("1234", 10),
      created_at: new Date(),
    },
  },
  { upsert: true },
);
database.admin_sessions.createIndex({ token_hash: 1 }, { unique: true });
database.admin_sessions.createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 });
print("Admin account and session indexes seeded");
