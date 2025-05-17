import React, { useState, useEffect } from 'react';

const TaskStatus = ({ taskId }) => {
  const [status, setStatus] = useState(null);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const response = await fetch(`/api/tasks/${taskId}`);
        if (!response.ok) throw new Error('Failed to fetch status');
        const data = await response.json();
        setStatus(data);
        setError('');
      } catch (err) {
        console.error('Fetch error:', err);
        // Only set error if the task is completed and still failing
        if (status && status.state === 'completed') {
          setError('Failed to fetch status');
        }
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 2000);
    return () => clearInterval(interval);
  }, [taskId]);

  if (!status) return null;

  return (
    <div className="mt-6 p-4 bg-gray-100 rounded-lg">
      <h2 className="text-xl font-bold mb-4">Task Status</h2>
      {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
      <p>Task ID: {status.id}</p>
      <p>State: {status.state}</p>
      {status.state === 'completed' && status.testsResults && (
        <p>Tests Passed: {status.testsResults.filter(r => r.successful).length} / {status.testsResults.length}</p>
      )}
    </div>
  );
};

export default TaskStatus;