import { useState } from 'react';
import type { Subscription } from '../types/aktuell';
import { Plus, Trash2, Database, Table } from 'lucide-react';

interface SubscriptionManagerProps {
  subscriptions: Subscription[];
  onSubscribe: (database: string, collection: string, filter?: Record<string, unknown>) => void;
  onUnsubscribe: (id: string) => void;
  disabled?: boolean;
}

export function SubscriptionManager({ 
  subscriptions, 
  onSubscribe, 
  onUnsubscribe, 
  disabled = false 
}: SubscriptionManagerProps) {
  const [database, setDatabase] = useState('');
  const [collection, setCollection] = useState('');
  const [filter, setFilter] = useState('{}');
  const [showForm, setShowForm] = useState(false);
  const [filterError, setFilterError] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!database.trim() || !collection.trim()) {
      return;
    }

    let parsedFilter: Record<string, unknown> | undefined;
    
    if (filter.trim() && filter.trim() !== '{}') {
      try {
        parsedFilter = JSON.parse(filter);
        setFilterError('');
      } catch {
        setFilterError('Invalid JSON filter');
        return;
      }
    }

    onSubscribe(database.trim(), collection.trim(), parsedFilter);
    
    // Reset form
    setDatabase('');
    setCollection('');
    setFilter('{}');
    setShowForm(false);
  };

  const formatFilter = (filter?: Record<string, unknown>) => {
    if (!filter || Object.keys(filter).length === 0) {
      return 'No filter';
    }
    return JSON.stringify(filter, null, 2);
  };

  return (
    <div className="bg-slate-800 rounded-lg border border-slate-700">
      <div className="p-4 border-b border-slate-700">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-100 flex items-center">
            <Database className="w-5 h-5 mr-2" />
            Subscriptions
          </h2>
          <button
            onClick={() => setShowForm(!showForm)}
            disabled={disabled}
            className="px-3 py-1 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-600 disabled:cursor-not-allowed text-white rounded-md text-sm transition-colors flex items-center"
          >
            <Plus className="w-4 h-4 mr-1" />
            Add
          </button>
        </div>
      </div>

      {/* Add Subscription Form */}
      {showForm && (
        <div className="p-4 border-b border-slate-700 bg-slate-750">
          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium mb-1">Database</label>
                <input
                  type="text"
                  value={database}
                  onChange={(e) => setDatabase(e.target.value)}
                  placeholder="database_name"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Collection</label>
                <input
                  type="text"
                  value={collection}
                  onChange={(e) => setCollection(e.target.value)}
                  placeholder="collection_name"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  required
                />
              </div>
            </div>
            
            <div>
              <label className="block text-sm font-medium mb-1">
                Filter (JSON) - Optional
              </label>
              <textarea
                value={filter}
                onChange={(e) => {
                  setFilter(e.target.value);
                  setFilterError('');
                }}
                placeholder='{"status": "active"}'
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm font-mono"
                rows={2}
              />
              {filterError && (
                <p className="text-red-400 text-xs mt-1">{filterError}</p>
              )}
            </div>
            
            <div className="flex space-x-2">
              <button
                type="submit"
                className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md text-sm transition-colors"
              >
                Subscribe
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForm(false);
                  setDatabase('');
                  setCollection('');
                  setFilter('{}');
                  setFilterError('');
                }}
                className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md text-sm transition-colors"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Subscription List */}
      <div className="max-h-80 overflow-y-auto">
        {subscriptions.length === 0 ? (
          <div className="p-4 text-center text-slate-400">
            <Database className="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>No active subscriptions</p>
            <p className="text-sm">Click "Add" to create your first subscription</p>
          </div>
        ) : (
          <div className="divide-y divide-slate-700">
            {subscriptions.map((sub) => (
              <div key={sub.id} className="p-3 hover:bg-slate-750 transition-colors">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center text-sm font-medium text-slate-100 mb-1">
                      <Table className="w-4 h-4 mr-1 text-blue-400" />
                      {sub.database}.{sub.collection}
                    </div>
                    <div className="text-xs text-slate-400 mb-2">
                      Created: {new Date(sub.createdAt).toLocaleString()}
                    </div>
                    <div className="text-xs">
                      <div className="text-slate-500 font-medium mb-1">Filter:</div>
                      <pre className="text-slate-400 bg-slate-900 p-2 rounded text-xs overflow-x-auto">
                        {formatFilter(sub.filter)}
                      </pre>
                    </div>
                  </div>
                  <button
                    onClick={() => onUnsubscribe(sub.id)}
                    disabled={disabled}
                    className="ml-2 p-1 text-slate-400 hover:text-red-400 disabled:cursor-not-allowed transition-colors"
                    title="Unsubscribe"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}