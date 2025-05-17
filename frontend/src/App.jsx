import React, { useState } from 'react';
import TaskForm from './components/TaskForm';
import TaskStatus from './components/TaskStatus';
import TaskList from './components/TaskList';

const App = () => {
  const [showStatus, setShowStatus] = useState(false);
  const [currentTaskId, setCurrentTaskId] = useState('');

  return (
    <div className="min-h-screen bg-gray-100 py-6 flex flex-col justify-center sm:py-12">
      <div className="relative py-3 sm:max-w-xl sm:mx-auto">
        <h1 className="text-2xl font-bold text-center mb-6">Code Runner</h1>
        <TaskForm setShowStatus={setShowStatus} setCurrentTaskId={setCurrentTaskId} />
        {showStatus && <TaskStatus taskId={currentTaskId} />}
        <TaskList />
      </div>
    </div>
  );
};

export default App;