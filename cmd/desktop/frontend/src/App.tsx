// frontend/src/App.tsx
import { useState, useEffect } from 'react'
import { Greet, Version } from '@wailsjs/go/ui/App'
import './App.css'

function App() {
  const [name, setName] = useState('')
  const [result, setResult] = useState('')
  const [version, setVersion] = useState('Loading...')

  useEffect(() => {
    Version().then((v: string) => setVersion(v))
  }, [])

  function doGreet() {
    Greet(name).then(setResult)
  }

  return (
    <div className="App">
      <div className="container">
        <h1>My App v{version}</h1>
        <div className="input-group">
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter your name"
          />
          <button onClick={doGreet}>Greet</button>
        </div>
        {result && <div className="result">{result}</div>}
      </div>
    </div>
  )
}

export default App
