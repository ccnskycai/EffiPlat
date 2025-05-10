import React, { useState } from 'react';

function App() {
  const [count, setCount] = useState(0);

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center">
      <header className="text-center">
        {/* <img src="/vite.svg" className="logo h-24 w-24 mx-auto mb-4" alt="Vite logo" /> */}
        <h1 className="text-4xl font-bold text-blue-600 mb-8">Vite + React</h1>
        <div className="card p-8 bg-white rounded-lg shadow-md">
          <button
            onClick={() => setCount((count) => count + 1)}
            className="px-6 py-2 bg-indigo-600 text-white font-semibold rounded-lg shadow-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-75"
          >
            count is {count}
          </button>
          <p className="mt-6 text-gray-600">
            Edit <code>src/App.tsx</code> and save to test HMR
          </p>
        </div>
        <p className="read-the-docs mt-8 text-gray-500">
          Click on the Vite and React logos to learn more
        </p>
      </header>
    </div>
  );
}

export default App;
