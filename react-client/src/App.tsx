import { useState, useEffect } from 'react'
import { Database, Zap, AlertCircle, Table } from 'lucide-react'
import { DataTable } from './components/DataTable'
import { SnapshotManager } from './components/SnapshotManager'
import type { ChangeEvent, SnapshotOptions } from './types/aktuell'
import { useAktuellStream } from './hooks/useAktuellStream'

interface ChangeDisplay extends ChangeEvent {
  id: string
  timestamp: string
}

function App() {
  const [serverUrl, setServerUrl] = useState('ws://localhost:8080/ws')
  
  const { 
    connectionStatus, 
    subscribe, 
    unsubscribe,
    subscription,
    snapshotData,
    snapshotStatus,
    clearSnapshotData,
    latestChange
  } = useAktuellStream({ url: serverUrl, snapshotOptions:{include_snapshot: true} })
  
  const [changes, setChanges] = useState<ChangeDisplay[]>([])
  const [activeTab, setActiveTab] = useState<'snapshot' | 'changes'>('snapshot')

  // Handle new changes from the stream
  useEffect(() => {
    if (latestChange) {
      const changeWithId: ChangeDisplay = {
        ...latestChange,
        id: `${latestChange.database}.${latestChange.collection}.${Date.now()}.${Math.random()}`,
        timestamp: new Date().toLocaleTimeString()
      }
      setChanges(prev => [changeWithId, ...prev.slice(0, 49)]) // Keep last 50 changes
    }
  }, [latestChange])

  const handleSubscribe = async (database: string, collection: string, filter?: Record<string, unknown>, snapshotOptions?: SnapshotOptions) => {
    try {
      console.log('Subscribing with options:', { database, collection, filter, snapshotOptions })
      subscribe(database, collection, filter, snapshotOptions)
      console.log('Subscribed successfully')
    } catch (err) {
      console.error('Subscription failed:', err)
    }
  }

  const handleUnsubscribe = (subscriptionId: string) => {
    unsubscribe(subscriptionId)
    setChanges([])
    clearSnapshotData()
  }


  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800 text-white p-4">
      <div className="max-w-7xl mx-auto">
        <header className="text-center mb-8">
          <div className="flex items-center justify-center gap-3 mb-4">
            <Database className="h-10 w-10 text-blue-400" />
            <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent">
              Aktuell Client
            </h1>
          </div>
          <p className="text-slate-300 text-lg">
            Real-time MongoDB Change Stream Visualization with Snapshots
          </p>
        </header>

        {/* Server URL Configuration */}
        <div className="mb-6 p-4 bg-slate-800 rounded-lg border border-slate-700">
          <label className="block text-sm font-medium mb-2">Server URL</label>
          <div className="flex gap-2">
            <input
              type="text"
              value={serverUrl}
              onChange={(e) => setServerUrl(e.target.value)}
              className="flex-1 px-3 py-2 bg-slate-700 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="ws://localhost:8080/ws"
            />
          </div>
        </div>

        {/* Connection Status */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4 mb-6">
          <div className="flex items-center gap-3">
            <div className={`w-3 h-3 rounded-full ${
              connectionStatus.connected ? 'bg-green-400' : 'bg-red-400'
            }`}></div>
            <span className="font-medium">
              {connectionStatus.connected ? 'Connected to Aktuell server' : 
               connectionStatus.connecting ? 'Connecting...' : 'Disconnected'}
            </span>
            {connectionStatus.error && (
              <div className="flex items-center gap-2 text-red-400 ml-auto">
                <AlertCircle className="h-4 w-4" />
                <span className="text-sm">{connectionStatus.error}</span>
              </div>
            )}
          </div>
        </div>

        {/* Subscription Manager */}
        <div className="mb-6">
          <SnapshotManager
            onSubscribe={handleSubscribe}
            snapshotStatus={snapshotStatus}
            onClearSnapshot={clearSnapshotData}
          />
        </div>

        {/* Current Subscription Status */}
        {subscription && (
          <div className="bg-slate-800 rounded-lg border border-slate-700 p-4 mb-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                <span className="font-medium">Active subscription: {subscription.database}.{subscription.collection}</span>
              </div>
              <button
                onClick={() => handleUnsubscribe(subscription.id)}
                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
              >
                Unsubscribe
              </button>
            </div>
          </div>
        )}

        {/* Data Display Tabs */}
        {subscription && (
          <div className="bg-slate-800 rounded-lg border border-slate-700 overflow-hidden">
            <div className="flex border-b border-slate-700">
              <button
                onClick={() => setActiveTab('snapshot')}
                className={`flex-1 px-6 py-4 font-medium transition-colors flex items-center justify-center gap-2 ${
                  activeTab === 'snapshot'
                    ? 'bg-slate-700 text-blue-400 border-b-2 border-blue-400'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                <Table className="h-4 w-4" />
                Snapshot Data ({snapshotData.length})
                {snapshotStatus.loading && <div className="w-2 h-2 bg-blue-400 rounded-full animate-pulse ml-1"></div>}
              </button>
              <button
                onClick={() => setActiveTab('changes')}
                className={`flex-1 px-6 py-4 font-medium transition-colors flex items-center justify-center gap-2 ${
                  activeTab === 'changes'
                    ? 'bg-slate-700 text-blue-400 border-b-2 border-blue-400'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                <Zap className="h-4 w-4" />
                Live Changes ({changes.length})
              </button>
            </div>

            <div className="p-6">
              {activeTab === 'snapshot' && (
                <div>
                  <div className="mb-4 p-3 bg-slate-700 rounded text-sm">
                    <strong>Debug Info:</strong>
                    <br />• Snapshot Data Length: {snapshotData.length}
                    <br />• Snapshot Status: {JSON.stringify(snapshotStatus)}
                    <br />• First Document: {snapshotData.length > 0 ? JSON.stringify(snapshotData[0]) : 'None'}
                  </div>
                  {snapshotData.length > 0 ? (
                    <DataTable
                      data={snapshotData}
                      title="Collection Snapshot"
                    />
                  ) : (
                    <div className="text-center py-8">
                      <Table className="h-12 w-12 text-slate-600 mx-auto mb-3" />
                      <p className="text-slate-400">No snapshot data available</p>
                      <p className="text-slate-500 text-sm mt-1">
                        Enable snapshot in subscription options to see existing data
                      </p>
                      {snapshotStatus.loading && (
                        <p className="text-blue-400 text-sm mt-2">Loading snapshot data...</p>
                      )}
                      {snapshotStatus.error && (
                        <p className="text-red-400 text-sm mt-2">Error: {snapshotStatus.error}</p>
                      )}
                    </div>
                  )}
                </div>
              )}

              {activeTab === 'changes' && (
                <div>
                  {changes.length > 0 ? (
                    <div className="space-y-3 max-h-96 overflow-y-auto">
                      {changes.map((change) => (
                        <div
                          key={change.id}
                          className="bg-slate-700 rounded-md p-4 border-l-4 border-blue-500"
                        >
                          <div className="flex items-center justify-between mb-2">
                            <span className={`px-2 py-1 rounded text-xs font-medium ${
                              change.operationType === 'insert' ? 'bg-green-600 text-white' :
                              change.operationType === 'update' ? 'bg-yellow-600 text-white' :
                              change.operationType === 'delete' ? 'bg-red-600 text-white' :
                              'bg-slate-600 text-white'
                            }`}>
                              {change.operationType.toUpperCase()}
                            </span>
                            <span className="text-slate-400 text-sm">{change.timestamp}</span>
                          </div>
                          
                          <div className="text-sm">
                            <div className="text-slate-300 mb-1">
                              <span className="font-medium">Collection:</span> {change.collection}
                            </div>
                            {change.documentKey && (
                              <div className="text-slate-300 mb-1">
                                <span className="font-medium">Document ID:</span> {JSON.stringify(change.documentKey)}
                              </div>
                            )}
                            {change.fullDocument && (
                              <div>
                                <div className="font-medium text-slate-300 mb-1">Full Document:</div>
                                <pre className="bg-slate-800 p-2 rounded text-xs overflow-x-auto">
                                  {JSON.stringify(change.fullDocument, null, 2)}
                                </pre>
                              </div>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <Zap className="h-12 w-12 text-slate-600 mx-auto mb-3" />
                      <p className="text-slate-400">No live changes detected</p>
                      <p className="text-slate-500 text-sm mt-1">
                        Changes will appear here when documents are modified
                      </p>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Empty State */}
        {snapshotData.length === 0 && changes.length === 0 && connectionStatus.connected && (
          <div className="text-center py-12">
            <Database className="h-16 w-16 text-slate-600 mx-auto mb-4" />
            <h3 className="text-xl font-medium text-slate-400 mb-2">Ready to stream data</h3>
            <p className="text-slate-500">
              Configure and subscribe to a collection to see snapshot data and live changes
            </p>
          </div>
        )}
      </div>
    </div>
  )
}

export default App
