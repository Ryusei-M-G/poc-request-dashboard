
const Home = () => {
  const clickHandle = () => {
    window.location.href = 'http://localhost:8080/login'
  }
  return (
    <div
      style={{ textAlign: 'center' }}>
      <div>ログイン</div>
      <button onClick={clickHandle}>ログイン with cognito</button>
    </div>
  )
}

export default Home