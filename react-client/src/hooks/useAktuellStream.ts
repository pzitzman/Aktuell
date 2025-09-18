import { useState, useEffect, useCallback, useRef } from 'react';
import type { ChangeEvent, ClientMessage, ServerMessage, ConnectionStatus, Subscription, SnapshotOptions } from '../types/aktuell';

interface UseAktuellOptions {
  url: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  snapshotOptions?: SnapshotOptions;
}
interface SnapshotStatus {
    loading: boolean;
    completed: boolean;
    error: string | null;
    errorCode: number | null;
    batch: number;
    remaining: number;
    total: number;
  };

interface UseAktuellReturn {
  connectionStatus: ConnectionStatus;
  latestChange: ChangeEvent | null;
  changes: ChangeEvent[];
  subscription: Subscription | null;
  snapshotData: Record<string, unknown>[];
  snapshotStatus: SnapshotStatus;
  subscribe: (database: string, collection: string, filter?: Record<string, unknown>, snapshotOptions?: SnapshotOptions) => void;
  unsubscribe: (subscriptionId: string) => void;
  clearChanges: () => void;
  clearSnapshotData: () => void;
  connect: () => void;
  disconnect: () => void;
  reconnect: () => void;
}

export const useAktuellStream = ({ 
  url, 
  reconnectInterval = 5000, 
  maxReconnectAttempts = 5,
  snapshotOptions = { include_snapshot: false} 
}: UseAktuellOptions): UseAktuellReturn => {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>({
    connected: false,
    connecting: false,
    reconnectAttempts: 0,
  });
  
  const [latestChange, setLatestChange] = useState<ChangeEvent | null>(null);
  const [changes, setChanges] = useState<ChangeEvent[]>([]);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [snapshotData, setSnapshotData] = useState<Record<string, unknown>[]>([]);
  const [snapshotStatus, setSnapshotStatus] = useState<SnapshotStatus>({
    loading: false,
    completed: false,
    error: null,
    errorCode: null,
    batch: 0,
    remaining: 0,
    total: 0,
  });
  
  const generateRequestId = () => `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const isManualDisconnect = useRef(false);
  const subscriptionsRef = useRef<Subscription[]>([]);
  const clientId = useRef(generateRequestId());

  // Clear any existing reconnect timeout
  const clearReconnectTimeout = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
  }, []);

  // Calculate reconnect delay with exponential backoff
  const getReconnectDelay = useCallback((attemptNumber: number) => {
    const baseDelay = reconnectInterval;
    const maxDelay = 30000; // Max 30 seconds
    const delay = Math.min(baseDelay * Math.pow(2, attemptNumber), maxDelay);
    return delay + Math.random() * 1000; // Add jitter
  }, [reconnectInterval]);

  // Attempt to reconnect with exponential backoff
  const attemptReconnect = useCallback(() => {
    if (isManualDisconnect.current) {
      console.log('🚫 Skipping reconnect - manual disconnect');
      return;
    }

    setConnectionStatus(prev => {
      const newAttemptCount = prev.reconnectAttempts + 1;
      
      if (newAttemptCount > maxReconnectAttempts) {
        console.log('❌ Max reconnect attempts reached:', maxReconnectAttempts);
        return {
          ...prev,
          connecting: false,
          error: `Failed to reconnect after ${maxReconnectAttempts} attempts`,
        };
      }

      console.log(`🔄 Attempting reconnect ${newAttemptCount}/${maxReconnectAttempts}...`);
      
      const delay = getReconnectDelay(newAttemptCount - 1);
      console.log(`⏰ Next reconnect in ${Math.round(delay / 1000)}s`);
      
      reconnectTimeoutRef.current = setTimeout(() => {
        // We'll trigger reconnect by calling connect from the component
        setConnectionStatus(curr => ({ ...curr, connecting: true }));
      }, delay);

      return {
        ...prev,
        connecting: true,
        reconnectAttempts: newAttemptCount,
        error: undefined,
      };
    });
  }, [maxReconnectAttempts, getReconnectDelay]);

  // Function to apply change events to snapshot data
  const applyChangeToSnapshot = useCallback((change: ChangeEvent) => {
    setSnapshotData(prev => {
      const newData = [...prev];
      
      if (change.operationType === 'insert' && change.fullDocument) {
        // Add new document for insert operations
        console.log('📝 Adding new document to snapshot:', change.fullDocument);
        newData.push(change.fullDocument);
      } else if (change.operationType === 'update' && change.fullDocument) {
        // Update existing document for update operations
        const index = newData.findIndex(doc => 
          JSON.stringify(doc._id) === JSON.stringify(change.documentKey._id)
        );
        if (index !== -1) {
          console.log('🔄 Updating document in snapshot at index:', index);
          newData[index] = change.fullDocument;
        } else {
          console.log('⚠️ Could not find document to update, adding as new');
          newData.push(change.fullDocument);
        }
      } else if (change.operationType === 'delete') {
        // Remove document for delete operations
        const index = newData.findIndex(doc => 
          JSON.stringify(doc._id) === JSON.stringify(change.documentKey._id)
        );
        if (index !== -1) {
          console.log('🗑️ Removing document from snapshot at index:', index);
          newData.splice(index, 1);
        }
      } else if (change.operationType === 'replace' && change.fullDocument) {
        // Replace document for replace operations
        const index = newData.findIndex(doc => 
          JSON.stringify(doc._id) === JSON.stringify(change.documentKey._id)
        );
        if (index !== -1) {
          console.log('🔁 Replacing document in snapshot at index:', index);
          newData[index] = change.fullDocument;
        } else {
          console.log('⚠️ Could not find document to replace, adding as new');
          newData.push(change.fullDocument);
        }
      }
      
      console.log('📊 Snapshot data updated, new count:', newData.length);
      return newData;
    });
  }, []);

  const sendMessage = useCallback((message: ClientMessage) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const messageJson = JSON.stringify(message);
      console.log('📤 Sending WebSocket message:', messageJson);
      wsRef.current.send(messageJson);
    } else {
      console.error('❌ Cannot send message - WebSocket not open. ReadyState:', wsRef.current?.readyState);
    }
  }, []);

  // Restore all active subscriptions after reconnection
  const restoreSubscriptions = useCallback(() => {
    console.log('🔄 Restoring subscriptions:', subscriptionsRef.current.length);
    
    subscriptionsRef.current.forEach(subscription => {
      const message: ClientMessage = {
        type: 'subscribe',
        database: subscription.database,
        collection: subscription.collection,
        requestId: subscription.id,
        snapshot_options: subscription.snapshot_options,
      };
      
      console.log('📤 Restoring subscription:', message);
      sendMessage(message);
    });
  }, [sendMessage]);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setConnectionStatus(prev => {
      if (prev.connecting) return prev;
      return { ...prev, connecting: true, error: undefined, reconnectAttempts: 0 };
    });

    isManualDisconnect.current = false;

    try {
      console.log('🚀 Creating new WebSocket connection to:', url);
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('🔗 WebSocket connection opened successfully');
        setConnectionStatus({
          connected: true,
          connecting: false,
          reconnectAttempts: 0,
          lastConnected: new Date(),
        });
        
        // Restore subscriptions after successful reconnection
        if (subscriptionsRef.current.length > 0) {
          console.log('🔄 Restoring subscriptions after reconnect');
          restoreSubscriptions();
        }
      };

      ws.onmessage = (event) => {
        console.log('🌟 MESSAGE RECEIVED!', event.data);
        try {
          const message: ServerMessage = JSON.parse(event.data);
          console.log('📨 Parsed message type:', message.type);
          
          if (message.type === 'change' && message.change) {
            console.log('🔄 Processing change message:', message.change);
            setLatestChange(message.change);
            setChanges(prev => [message.change!, ...prev.slice(0, 99)]);
            // Apply the change to the snapshot data
            applyChangeToSnapshot(message.change);
          } else if (message.type === 'snapshot_start') {
            console.log('� Processing snapshot_start message');
            setSnapshotStatus(prev => ({
              ...prev,
              loading: true,
              completed: false,
              error: null,
              batch: 0,
            }));
          } else if (message.type === 'snapshot') {
            console.log('�📦 Snapshot message with', message.snapshot_data?.length, 'documents');
            if (message.snapshot_data) {
              setSnapshotData(prev => {
                const newData = [...prev, ...message.snapshot_data!];
                console.log('📊 Total snapshot data:', newData.length);
                return newData;
              });
              setSnapshotStatus(prev => ({
                ...prev,
                batch: message.snapshot_batch || prev.batch,
                remaining: message.snapshot_remaining || 0,
                total: message.snapshot_total || prev.total,
              }));
            }
          } else if (message.type === 'snapshot_end') {
            console.log('✅ Processing snapshot_end message');
            setSnapshotStatus(prev => ({
              ...prev,
              loading: false,
              completed: true,
            }));
          } else if (message.type === 'error') {
            console.error('❌ Error message:', message.error);
            setSnapshotStatus(prev => ({
              ...prev,
              loading: false,
              error: message.error || 'Unknown error',
              errorCode: message.errorCode || null,
            }));
          } else {
            console.log('🤔 Unhandled message type:', message.type);
          }
        } catch (err) {
          console.error('💥 Failed to parse message:', err);
        }
      };

      ws.onclose = (event) => {
        console.log('🔌 WebSocket closed - code:', event.code, 'reason:', event.reason, 'wasClean:', event.wasClean);
        setConnectionStatus(prev => ({ ...prev, connected: false, connecting: false }));
        
        // Only attempt reconnect if not a manual disconnect and not a clean close
        if (!isManualDisconnect.current && event.code !== 1000) {
          console.log('🔄 Connection lost unexpectedly, attempting reconnect...');
          attemptReconnect();
        } else if (isManualDisconnect.current) {
          console.log('🚫 Manual disconnect - no reconnect attempt');
        } else {
          console.log('✅ Clean disconnect - no reconnect needed');
        }
      };

      ws.onerror = (error) => {
        console.error('💥 WebSocket error:', error);
        setConnectionStatus(prev => ({
          ...prev,
          connected: false,
          connecting: false,
          error: 'Connection error occurred',
        }));
        
        // Don't immediately reconnect on error - let onclose handle it
      };

    } catch (error) {
      console.error('💥 Connection failed:', error);
      setConnectionStatus(prev => ({
        ...prev,
        connected: false,
        connecting: false,
        error: 'Failed to create WebSocket connection',
      }));
      
      // Attempt reconnect after connection creation failure
      if (!isManualDisconnect.current) {
        attemptReconnect();
      }
    }
  }, [url, applyChangeToSnapshot, attemptReconnect, restoreSubscriptions]);

  const disconnect = useCallback(() => {
    console.log('🔌 Manual disconnect requested');
    isManualDisconnect.current = true;
    clearReconnectTimeout();
    
    if (wsRef.current) {
      wsRef.current.close(1000, 'Manual disconnect'); // Clean close
      wsRef.current = null;
    }
    
    setConnectionStatus({ connected: false, connecting: false, reconnectAttempts: 0 });
    
    // Clear subscription on manual disconnect
    setSubscription(null);
    subscriptionsRef.current = [];
  }, [clearReconnectTimeout]);

  useEffect(() => {
    connect();
    return () => {
      disconnect();
      clearReconnectTimeout();
    };
  }, [connect, disconnect, clearReconnectTimeout]);

  // Effect to handle reconnection when connection status indicates it should reconnect
  useEffect(() => {
    if (connectionStatus.connecting && !connectionStatus.connected && connectionStatus.reconnectAttempts > 0) {
      console.log('🔄 Triggering reconnection attempt...');
      connect();
    }
  }, [connectionStatus.connecting, connectionStatus.connected, connectionStatus.reconnectAttempts, connect]);

  const subscribe = useCallback((database: string, collection: string, ) => {
    console.log('Subscribing:', { database, collection, snapshotOptions });
    
    setSnapshotData([]);
    
    const subscription: Subscription = {
      id: generateRequestId(),
      clientId: clientId.current,
      database,
      collection,
      snapshot_options: snapshotOptions,
      createdAt: new Date().toISOString(),
    };

    const message: ClientMessage = {
      type: 'subscribe',
      database,
      collection,
      requestId: subscription.id,
      snapshot_options: snapshotOptions,
    };

    sendMessage(message);
    
    // Store subscription in both state and ref for restoration
    setSubscription(subscription);
    subscriptionsRef.current = [subscription];
  }, [sendMessage, snapshotOptions]);

  const unsubscribe = useCallback((subscriptionId: string) => {
    const message: ClientMessage = { type: 'unsubscribe', requestId: subscriptionId };
    sendMessage(message);
    
    // Clear subscriptions from both state and ref
    setSubscription(null);
    subscriptionsRef.current = [];
  }, [sendMessage]);

  const clearChanges = useCallback(() => {
    setChanges([]);
    setLatestChange(null);
  }, []);

  const clearSnapshotData = useCallback(() => {
    setSnapshotData([]);
    setSnapshotStatus({
      loading: false,
      completed: false,
      error: null,
      batch: 0,
      remaining: 0,
      total: 0,
      errorCode: null
    });
  }, []);

  const reconnect = useCallback(() => {
    console.log('🔄 Manual reconnect requested');
    setConnectionStatus(prev => ({
      ...prev,
      reconnectAttempts: 0,
      error: undefined,
    }));
    clearReconnectTimeout();
    disconnect();
    setTimeout(() => connect(), 100); // Small delay to ensure clean disconnect
  }, [clearReconnectTimeout, disconnect, connect]);

  return {
    connectionStatus,
    latestChange,
    changes,
    subscription,
    snapshotData,
    snapshotStatus,
    subscribe,
    unsubscribe,
    clearChanges,
    clearSnapshotData,
    connect,
    disconnect,
    reconnect,
  };
};
