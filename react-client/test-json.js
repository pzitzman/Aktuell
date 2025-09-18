// Quick test to see what JSON structure we're sending
// Example usage with Aktuell types

const snapshotOptions = {
  include_snapshot: true,
  snapshot_limit: 50,
  batch_size: 10,
  snapshot_sort: { _id: 1 }
};

const message = {
  type: 'subscribe',
  database: 'aktuell',
  collection: 'users',
  requestId: 'test-123',
  snapshot_options: snapshotOptions
};

console.log('JSON that would be sent to server:');
console.log(JSON.stringify(message, null, 2));