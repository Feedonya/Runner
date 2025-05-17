import React, { useState } from 'react';
import TaskForm from './components/TaskForm.jsx';
import TaskStatus from './components/TaskStatus.jsx';
import TaskList from './components/TaskList.jsx';

function App() {
  const [selectedTaskId, setSelectedTaskId] = useState(null);

  return (
    <div className="min-h-screen bg-gray-100 p-6">
      <h1 className="text-3xl font-bold text-center mb-8">Code Runner</h1>
      <div className="max-w-4xl mx-auto space-y-8">
        <TaskForm onTaskSubmitted={setSelectedTaskId} />
        {selectedTaskId && <TaskStatus taskId={selectedTaskId} />}
        <TaskList />
      </div>
    </div>
  );
}

export default App;