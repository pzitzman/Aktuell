import { useState } from 'react';
import { Camera, Settings } from 'lucide-react';
import type { SnapshotOptions } from '../types/aktuell';

interface SnapshotManagerProps {
  onSubscribe: (database: string, collection: string, filter?: Record<string, unknown>, snapshotOptions?: SnapshotOptions) => void;
  snapshotStatus: {
    loading: boolean;
    completed: boolean;
    error: string | null;
    batch: number;
    remaining: number;
    total: number;
  };
  onClearSnapshot: () => void;
}

export const SnapshotManager = ({ onSubscribe, snapshotStatus, onClearSnapshot }: SnapshotManagerProps) => {
  const [database, setDatabase] = useState('aktuell');
  const [collection, setCollection] = useState('users');
  const [includeSnapshot, setIncludeSnapshot] = useState(true);
  const [snapshotLimit, setSnapshotLimit] = useState(50);
  const [batchSize, setBatchSize] = useState(10);
  const [showAdvanced, setShowAdvanced] = useState(false);

  const handleSubscribe = () => {
    const snapshotOptions: SnapshotOptions = {
      include_snapshot: includeSnapshot,
      snapshot_limit: snapshotLimit > 0 ? snapshotLimit : undefined,
      batch_size: batchSize > 0 ? batchSize : undefined,
      snapshot_sort: { _id: 1 }, // Sort by _id ascending
    };

    onSubscribe(database, collection, undefined, includeSnapshot ? snapshotOptions : undefined);
  };

  return (
    <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
      <div className="flex items-center gap-2 mb-4">
        <Camera className="h-5 w-5 text-blue-400" />
        <h3 className="text-lg font-semibold">Snapshot Subscription</h3>
      </div>

      <div className="space-y-4">
        {/* Database and Collection */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">Database</label>
            <input
              type="text"
              value={database}
              onChange={(e) => setDatabase(e.target.value)}
              className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="aktuell"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Collection</label>
            <input
              type="text"
              value={collection}
              onChange={(e) => setCollection(e.target.value)}
              className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="users"
            />
          </div>
        </div>

        {/* Snapshot Options */}
        <div className="flex items-center gap-2">
          <input
            id="include-snapshot"
            type="checkbox"
            checked={includeSnapshot}
            onChange={(e) => setIncludeSnapshot(e.target.checked)}
            className="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded focus:ring-blue-500 focus:ring-2"
          />
          <label htmlFor="include-snapshot" className="text-sm font-medium">
            Include initial snapshot (existing data)
          </label>
        </div>

        {includeSnapshot && (
          <div className="pl-6 border-l-2 border-blue-500/20 space-y-3">
            <div className="flex items-center gap-2">
              <button
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
              >
                <Settings className="h-4 w-4" />
                {showAdvanced ? 'Hide' : 'Show'} Advanced Options
              </button>
            </div>

            {showAdvanced && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Snapshot Limit
                    <span className="text-slate-400 text-xs ml-1">(max documents)</span>
                  </label>
                  <input
                    type="number"
                    value={snapshotLimit}
                    onChange={(e) => setSnapshotLimit(Math.max(1, parseInt(e.target.value) || 50))}
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    min="1"
                    max="10000"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Batch Size
                    <span className="text-slate-400 text-xs ml-1">(documents per batch)</span>
                  </label>
                  <input
                    type="number"
                    value={batchSize}
                    onChange={(e) => setBatchSize(Math.max(1, parseInt(e.target.value) || 10))}
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    min="1"
                    max="1000"
                  />
                </div>
              </div>
            )}
          </div>
        )}

        {/* Snapshot Status */}
        {(snapshotStatus.loading || snapshotStatus.completed || snapshotStatus.error) && (
          <div className="p-3 bg-slate-700 rounded-md border border-slate-600">
            <div className="text-sm font-medium mb-1">Snapshot Status</div>
            {snapshotStatus.loading && (
              <div className="flex items-center gap-2 text-blue-400">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-400"></div>
                <span>Loading batch {snapshotStatus.batch}...</span>
                {snapshotStatus.remaining > 0 && (
                  <span className="text-slate-400">({snapshotStatus.remaining} remaining)</span>
                )}
              </div>
            )}
            {snapshotStatus.completed && (
              <div className="text-green-400">✓ Snapshot completed successfully</div>
            )}
            {snapshotStatus.error && (
              <div className="text-red-400">✗ Error: {snapshotStatus.error}</div>
            )}
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex gap-2">
          <button
            onClick={handleSubscribe}
            disabled={!database || !collection || snapshotStatus.loading}
            className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-600 disabled:cursor-not-allowed text-white rounded-md transition-colors"
          >
            Subscribe{includeSnapshot ? ' with Snapshot' : ''}
          </button>
          {(snapshotStatus.completed || snapshotStatus.error) && (
            <button
              onClick={onClearSnapshot}
              className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md transition-colors"
            >
              Clear
            </button>
          )}
        </div>
      </div>
    </div>
  );
};