import React, { useState, useEffect } from 'react';

function TaskList() {
  const [tasks, setTasks] = useState([]);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const response = await fetch('http://backend:8080/api/tasks');
        if (!response.ok) {
          throw new Error('Failed to fetch tasks');
        }
        const data = await response.json();
        setTasks(data);
      } catch (err) {
        setError(err.message);
      }
    };

    fetchTasks();
  }, []);

  if (error) {
    return <p className="text-red-500">{error}</p>;
  }

  return (
    <div className="bg-white p-6 rounded-lg shadow-md">
      <h2 className="text-xl font-semibold mb-4">Submitted Tasks</h2>
      {tasks.length === 0 ? (
        <p className="text-gray-500">No tasks submitted yet.</p>
      ) : (
        <ul className="space-y-2">
          {tasks.map(task => {
            const passedTests = task.testsResults ? task.testsResults.filter(r => r.successful).length : 0;
            const totalTests = task.testsResults ? task.testsResults.length : 0;
            return (
              <li key={task.id} className="border-b py-2">
                <p><strong>Task ID:</strong> {task.id}</p>
                <p><strong>State:</strong> {task.state}</p>
                {task.testsResults && (
                  <p><strong>Tests Passed:</strong> {passedTests} / {totalTests}</p>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

export default TaskList;