import { useState } from 'react';
import type { ChangeEvent } from '../types/aktuell';
import { Trash2, FileText, Calendar, Database, Table, ChevronDown, ChevronRight } from 'lucide-react';

interface ChangeEventsListProps {
  changes: ChangeEvent[];
  onClear: () => void;
}

export function ChangeEventsList({ changes, onClear }: ChangeEventsListProps) {
  const [expandedEvents, setExpandedEvents] = useState<Set<string>>(new Set());

  const toggleExpanded = (eventId: string) => {
    const newExpanded = new Set(expandedEvents);
    if (newExpanded.has(eventId)) {
      newExpanded.delete(eventId);
    } else {
      newExpanded.add(eventId);
    }
    setExpandedEvents(newExpanded);
  };

  const getOperationColor = (operationType: string) => {
    switch (operationType) {
      case 'insert': return 'text-green-400 bg-green-900/20';
      case 'update': return 'text-blue-400 bg-blue-900/20';
      case 'delete': return 'text-red-400 bg-red-900/20';
      case 'replace': return 'text-yellow-400 bg-yellow-900/20';
      case 'drop': return 'text-purple-400 bg-purple-900/20';
      case 'rename': return 'text-orange-400 bg-orange-900/20';
      default: return 'text-slate-400 bg-slate-900/20';
    }
  };

  const formatJson = (obj: unknown) => {
    if (!obj) return 'null';
    return JSON.stringify(obj, null, 2);
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleString();
    } catch {
      return timestamp;
    }
  };

  return (
    <div className="bg-slate-800 rounded-lg border border-slate-700">
      <div className="p-4 border-b border-slate-700">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-100 flex items-center">
            <FileText className="w-5 h-5 mr-2" />
            Change Events ({changes.length})
          </h2>
          {changes.length > 0 && (
            <button
              onClick={onClear}
              className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white rounded-md text-sm transition-colors flex items-center"
            >
              <Trash2 className="w-4 h-4 mr-1" />
              Clear All
            </button>
          )}
        </div>
      </div>

      <div className="max-h-96 overflow-y-auto">
        {changes.length === 0 ? (
          <div className="p-8 text-center text-slate-400">
            <FileText className="w-16 h-16 mx-auto mb-4 opacity-50" />
            <p className="text-lg mb-2">No change events yet</p>
            <p className="text-sm">
              Subscribe to a database collection to see real-time changes
            </p>
          </div>
        ) : (
          <div className="divide-y divide-slate-700">
            {changes.slice().reverse().map((event, index) => {
              const isExpanded = expandedEvents.has(event.id);
              return (
                <div key={`${event.id}-${index}`} className="p-3">
                  <div 
                    className="flex items-start justify-between cursor-pointer hover:bg-slate-750 rounded p-2 transition-colors"
                    onClick={() => toggleExpanded(event.id)}
                  >
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        <span className={`px-2 py-1 rounded text-xs font-medium ${getOperationColor(event.operationType)}`}>
                          {event.operationType.toUpperCase()}
                        </span>
                        <div className="flex items-center text-sm text-slate-300">
                          <Database className="w-3 h-3 mr-1" />
                          {event.database}
                          <Table className="w-3 h-3 mx-1" />
                          {event.collection}
                        </div>
                      </div>
                      
                      <div className="flex items-center text-xs text-slate-400 space-x-3">
                        <div className="flex items-center">
                          <Calendar className="w-3 h-3 mr-1" />
                          {formatTimestamp(event.timestamp)}
                        </div>
                        <div className="text-slate-500">
                          ID: {event.id.slice(0, 8)}...
                        </div>
                      </div>
                    </div>
                    
                    <div className="ml-2">
                      {isExpanded ? (
                        <ChevronDown className="w-4 h-4 text-slate-400" />
                      ) : (
                        <ChevronRight className="w-4 h-4 text-slate-400" />
                      )}
                    </div>
                  </div>

                  {isExpanded && (
                    <div className="mt-3 space-y-3 bg-slate-900 rounded-lg p-3">
                      {/* Document Key */}
                      <div>
                        <h4 className="text-sm font-medium text-slate-300 mb-1">Document Key:</h4>
                        <pre className="text-xs bg-slate-800 p-2 rounded overflow-x-auto text-slate-400">
                          {formatJson(event.documentKey)}
                        </pre>
                      </div>

                      {/* Full Document (for inserts/replaces) */}
                      {event.fullDocument && (
                        <div>
                          <h4 className="text-sm font-medium text-slate-300 mb-1">Full Document:</h4>
                          <pre className="text-xs bg-slate-800 p-2 rounded overflow-x-auto text-slate-400 max-h-40 overflow-y-auto">
                            {formatJson(event.fullDocument)}
                          </pre>
                        </div>
                      )}

                      {/* Updated Fields (for updates) */}
                      {event.updatedFields && Object.keys(event.updatedFields).length > 0 && (
                        <div>
                          <h4 className="text-sm font-medium text-slate-300 mb-1">Updated Fields:</h4>
                          <pre className="text-xs bg-slate-800 p-2 rounded overflow-x-auto text-slate-400">
                            {formatJson(event.updatedFields)}
                          </pre>
                        </div>
                      )}

                      {/* Removed Fields (for updates) */}
                      {event.removedFields && event.removedFields.length > 0 && (
                        <div>
                          <h4 className="text-sm font-medium text-slate-300 mb-1">Removed Fields:</h4>
                          <pre className="text-xs bg-slate-800 p-2 rounded overflow-x-auto text-slate-400">
                            {formatJson(event.removedFields)}
                          </pre>
                        </div>
                      )}

                      {/* Timestamps */}
                      <div className="flex justify-between text-xs text-slate-500 pt-2 border-t border-slate-700">
                        <span>Server: {formatTimestamp(event.timestamp)}</span>
                        <span>Client: {formatTimestamp(event.clientTimestamp)}</span>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}