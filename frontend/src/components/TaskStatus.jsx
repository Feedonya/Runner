import React, { useState, useEffect } from 'react';

function TaskStatus({ taskId }) {
  const [task, setTask] = useState(null);
  const [error, setError] = useState('');

  useEffect(() => {
    const pollTask = async () => {
      try {
        const response = await fetch(`http://backend:8080/api/tasks/${taskId}`);
        if (!response.ok) {
          throw new Error('Failed to fetch task status');
        }
        const data = await response.json();
        setTask(data);
      } catch (err) {
        setError(err.message);
      }
    };

    pollTask();
    const interval = setInterval(pollTask, 2000);
    return () => clearInterval(interval);
  }, [taskId]);

  if (error) {
    return <p className="text-red-500">{error}</p>;
  }

  if (!task) {
    return <p className="text-gray-500">Loading task status...</p>;
  }

  const passedTests = task.testsResults ? task.testsResults.filter(r => r.successful).length : 0;
  const totalTests = task.testsResults ? task.testsResults.length : 0;

  return (
    <div className="bg-white p-6 rounded-lg shadow-md">
      <h2 className="text-xl font-semibold mb-4">Task Status: {taskId}</h2>
      <p><strong>State:</strong> {task.state}</p>
      {task.testsResults && (
        <p>
          <strong>Tests Passed:</strong> {passedTests} / {totalTests}
        </p>
      )}
    </div>
  );
}

export default TaskStatus;