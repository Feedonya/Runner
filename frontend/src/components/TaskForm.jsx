import React, { useState } from 'react';

const TaskForm = ({ setShowStatus, setCurrentTaskId }) => {
  const [taskId, setTaskId] = useState('');
  const [compiler, setCompiler] = useState('g++');
  const [codeFile, setCodeFile] = useState(null);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    console.log('Submitting:', { taskId, compiler, codeFile });
    if (!taskId || !compiler || !codeFile) {
      setError('Please fill all fields and select a file.');
      return;
    }

    try {
      const response = await fetch('/api/tasks', {
        method: 'POST',
        body: formData,
      });
      if (!response.ok) throw new Error('Failed to submit');
      const data = await response.json();
      console.log('Response:', data);
      setCurrentTaskId(data.task_id); // Use the prop instead of local setTaskId
      setShowStatus(true);
    } catch (err) {
      setError(err.message);
      console.error('Fetch error:', err);
    } finally {
      setSubmitting(false);
    }

    setError('');
    setSubmitting(true);

    const formData = new FormData();
    formData.append('task_id', taskId);
    formData.append('compiler', compiler);
    formData.append('code', codeFile);

    try {
      const response = await fetch('http://localhost:8080/api/tasks', {
        method: 'POST',
        body: formData,
      });
      if (!response.ok) throw new Error('Failed to submit');
      const data = await response.json();
      console.log('Response:', data);
      setCurrentTaskId(data.task_id); // Update local state
      setShowStatus(true); // Notify parent to show status
    } catch (err) {
      setError(err.message);
      console.error('Fetch error:', err);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="p-4 max-w-md mx-auto bg-white rounded-lg shadow-md">
      <h2 className="text-xl font-bold mb-4">Submit Solution</h2>
      <form onSubmit={handleSubmit}>
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700">Task ID</label>
          <input
            type="text"
            value={taskId}
            onChange={(e) => setTaskId(e.target.value)}
            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm"
            placeholder="task_123"
          />
        </div>
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700">Compiler</label>
          <select
            value={compiler}
            onChange={(e) => setCompiler(e.target.value)}
            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm"
          >
            <option value="g++">g++</option>
          </select>
        </div>
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700">Code File</label>
          <input
            type="file"
            accept=".cpp" // Restrict to .cpp files
            onChange={(e) => setCodeFile(e.target.files[0])}
            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm"
          />
        </div>
        {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
        <button
          type="submit"
          disabled={submitting}
          className="w-full bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700 disabled:bg-gray-400"
        >
          Submit
        </button>
      </form>
    </div>
  );
};

export default TaskForm;