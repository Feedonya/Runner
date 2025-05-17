import React, { useState, useEffect } from 'react';

const TaskStatus = ({ taskId }) => {
  const [status, setStatus] = useState(null);
  const [error, setError] = useState('');
  const [lastFetch, setLastFetch] = useState(null);

  useEffect(() => {
    let isMounted = true;

    const fetchStatus = async () => {
      try {
        const response = await fetch(`/api/tasks/${taskId}`);
        if (!response.ok) {
          if (response.status === 404) throw new Error('Task not found');
          throw new Error('Failed to fetch status');
        }
        const data = await response.json();
        if (isMounted) {
          setLastFetch(new Date().toISOString());
          if (data.state === 'completed') {
            if (!data.testsResults) {
              console.warn('Completed but missing testsResults at', lastFetch, '- retrying...');
            } else {
              setStatus(data);
              setError('');
            }
          } else {
            setStatus(data); // Update for intermediate states
            setError('');
          }
        }
      } catch (err) {
        console.error('Fetch error at', lastFetch, ':', err);
        if (isMounted && status && status.state === 'completed') {
          setError('Failed to fetch status after completion');
        }
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 2000); // Poll every 2 seconds
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, [taskId]);

  if (!status) return <div className="mt-6 p-4 bg-gray-100 rounded-lg">Loading status...</div>;

  return (
    <div className="mt-6 p-4 bg-gray-100 rounded-lg">
      <h2 className="text-xl font-bold mb-4">Task Status</h2>
      {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
      <p>Task ID: {status.id}</p>
      <p>State: {status.state}</p>
      <p>Last Fetch: {lastFetch}</p>
      {status.state === 'completed' && status.testsResults && (
        <p>Tests Passed: {status.testsResults.filter(r => r.passed).length} / {status.testsResults.length}</p>
      )}
    </div>
  );
};

export default TaskStatus;