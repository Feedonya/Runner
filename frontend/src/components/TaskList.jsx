import React, { useState, useEffect } from 'react';

const TaskList = () => {
  const [tasks, setTasks] = useState([]);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const response = await fetch('/api/tasks');
        if (!response.ok) throw new Error('Failed to fetch tasks');
        const data = await response.json();
        setTasks(data);
        setError(''); // Clear error on success
      } catch (err) {
        console.error('Fetch error:', err);
        // Retry after a delay instead of showing error immediately
        setTimeout(fetchTasks, 2000); // Retry after 2 seconds
      }
    };
    fetchTasks();
  }, []);

  return (
    <div className="mt-6">
      <h2 className="text-xl font-bold mb-4">Submitted Tasks</h2>
      {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
      <ul>
        {tasks.map((task) => (
          <li key={task.id} className="mb-2">
            {task.id} - {task.state}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default TaskList;