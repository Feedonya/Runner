import React, { useState } from 'react';

function TaskForm({ onTaskSubmitted }) {
  const [taskId, setTaskId] = useState('');
  const [compiler, setCompiler] = useState('g++');
  const [codeFile, setCodeFile] = useState(null);
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);

    if (!taskId || !codeFile) {
      setError('Task ID and code file are required');
      setIsSubmitting(false);
      return;
    }

    const formData = new FormData();
    formData.append('task_id', taskId);
    formData.append('compiler', compiler);
    formData.append('code', codeFile);

    try {
      const response = await fetch('http://backend:8080/api/tasks', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Failed to submit task');
      }

      const data = await response.json();
      onTaskSubmitted(data.task_id);
      setTaskId('');
      setCodeFile(null);
      setCompiler('g++');
    } catch (err) {
      setError(err.message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow-md">
      <h2 className="text-xl font-semibold mb-4">Submit Solution</h2>
      {error && <p className="text-red-500 mb-4">{error}</p>}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">Task ID</label>
          <input
            type="text"
            value={taskId}
            onChange={(e) => setTaskId(e.target.value)}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
            placeholder="e.g., task_123"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">Compiler</label>
          <select
            value={compiler}
            onChange={(e) => setCompiler(e.target.value)}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          >
            <option value="g++">g++</option>
            <option value="gcc">gcc</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">Code File</label>
          <input
            type="file"
            accept=".cpp"
            onChange={(e) => setCodeFile(e.target.files[0])}
            className="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
          />
        </div>
        <button
          type="submit"
          disabled={isSubmitting}
          className={`w-full py-2 px-4 rounded-md text-white ${
            isSubmitting ? 'bg-indigo-400 cursor-not-allowed' : 'bg-indigo-600 hover:bg-indigo-700'
          }`}
        >
          {isSubmitting ? 'Submitting...' : 'Submit'}
        </button>
      </form>
    </div>
  );
}

export default TaskForm;