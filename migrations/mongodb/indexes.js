// This script creates necessary indexes for the collections in the oauth2_provider database.
// Run this script using mongosh:
// mongosh "mongodb://root:password@localhost:27017" --file migrations/mongodb/indexes.js

// Select the database
db = db.getSiblingDB('oauth2_provider');

// --- Users Collection ---
// Create a unique index on the 'username' field for fast lookups and to prevent duplicates.
db.users.createIndex({ "username": 1 }, { unique: true });
print("Created index on users.username");

// --- Clients Collection ---
// Create a unique index on the 'client_id' field for fast lookups.
db.clients.createIndex({ "client_id": 1 }, { unique: true });
print("Created index on clients.client_id");

// --- Tokens Collection ---
// Create an index on 'signature' for refresh tokens for fast lookups.
db.tokens.createIndex({ "signature": 1 });
// Create a TTL index on 'expires_at' to automatically delete expired documents.
db.tokens.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 });
print("Created indexes on tokens collection");

print("Index creation complete.");